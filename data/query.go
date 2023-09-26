package data

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// SearchURLParams is a struct which contains all information of search url parameters and values
type SearchURLParams struct {
	Query                string
	PopulationTypeFilter string
	DimensionsFilter     string
	Filter               Filter
	TopicFilter          string
	Sort                 Sort
	Limit                int
	CurrentPage          int
	Offset               int
}

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if paginationErr != nil {
		log.Error(ctx, "unable to review pagination", paginationErr)
		return validatedQueryParams, paginationErr
	}

	reviewSort(ctx, cfg, urlQuery, &validatedQueryParams)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &validatedQueryParams)
	if contentTypeFilterError != nil {
		log.Error(ctx, "invalid content type filters set", contentTypeFilterError)
		return validatedQueryParams, contentTypeFilterError
	}
	topicFilterErr := reviewTopicFilters(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
	if topicFilterErr != nil {
		log.Error(ctx, "invalid topic filters set", topicFilterErr)
		return validatedQueryParams, topicFilterErr
	}
	populationTypeFilterErr := reviewPopulationTypeFilters(urlQuery, &validatedQueryParams)
	if populationTypeFilterErr != nil {
		log.Error(ctx, "invalid population types set", populationTypeFilterErr)
		return validatedQueryParams, populationTypeFilterErr
	}
	dimensionsFilterErr := reviewDimensionsFilters(urlQuery, &validatedQueryParams)
	if dimensionsFilterErr != nil {
		log.Error(ctx, "invalid population types set", dimensionsFilterErr)
		return validatedQueryParams, dimensionsFilterErr
	}

	queryStringErr := reviewQueryString(ctx, urlQuery)
	if queryStringErr == nil {
		return validatedQueryParams, nil
	} else if errors.Is(queryStringErr, apperrors.ErrInvalidQueryCharLengthString) && hasFilters(validatedQueryParams) {
		log.Info(ctx, "the query string did not pass review")
		return validatedQueryParams, nil
	}

	return validatedQueryParams, queryStringErr
}

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewDatasetQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if paginationErr != nil {
		log.Error(ctx, "unable to review pagination", paginationErr)
		return validatedQueryParams, paginationErr
	}

	reviewDatasetSort(ctx, cfg, urlQuery, &validatedQueryParams)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &validatedQueryParams)
	if contentTypeFilterError != nil {
		log.Error(ctx, "invalid content type filters set", contentTypeFilterError)
		return validatedQueryParams, contentTypeFilterError
	}
	topicFilterErr := reviewTopicFilters(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
	if topicFilterErr != nil {
		log.Error(ctx, "invalid topic filters set", topicFilterErr)
		return validatedQueryParams, topicFilterErr
	}
	populationTypeFilterErr := reviewPopulationTypeFilters(urlQuery, &validatedQueryParams)
	if populationTypeFilterErr != nil {
		log.Error(ctx, "invalid population types set", populationTypeFilterErr)
		return validatedQueryParams, populationTypeFilterErr
	}
	dimensionsFilterErr := reviewDimensionsFilters(urlQuery, &validatedQueryParams)
	if dimensionsFilterErr != nil {
		log.Error(ctx, "invalid population types set", dimensionsFilterErr)
		return validatedQueryParams, dimensionsFilterErr
	}

	return validatedQueryParams, nil
}

// GetSearchAPIQuery gets the query that needs to be passed to the search-api to get search results
func GetSearchAPIQuery(validatedQueryParams SearchURLParams, censusTopicCache *cache.Topic) url.Values {
	apiQuery := createSearchAPIQuery(validatedQueryParams)

	// update content_type query (filters) with sub filters
	updateQueryWithAPIFilters(apiQuery)

	// update topics query with sub topics for dp-search-api
	updateTopicsQueryForSearchAPI(apiQuery, censusTopicCache)

	return apiQuery
}

// GetDataAggregationQuery gets the query that needs to be passed to the search-api to get data aggregation results
func GetDataAggregationQuery(validatedQueryParams SearchURLParams, template string) url.Values {
	apiQuery := createSearchAPIQuery(validatedQueryParams)
	var contentTypes = ""
	switch template {
	case "all-adhocs":
		contentTypes = "static_adhoc"
	case "all-methodologies":
		contentTypes = "static_qmi," + "static_methodology," + "static_methodology_download"
	case "home-datalist":
		contentTypes = "static_adhoc," + "timeseries," + "dataset_landing_page," + "reference_tables"
	case "home-publications":
		contentTypes = "bulletin," + "article," + "article_download," + "compendium_landing_page"
	case "published-requests":
		contentTypes = "static_foi"
	case "home-list":
		contentTypes = "static_page"
	case "home-methodology":
		contentTypes = "static_qmi," + "static_methodology," + "static_methodology_download"
	case "time-series-tool":
		contentTypes = "timeseries"
		//compendia
	}

	log.Info(context.TODO(), "kur", log.Data{
		"apiqueryyyy": apiQuery.Get("content_type"),
	})
	// log.Info(context.TODO(), "kur", log.Data{
	// 	"apiqueryyyy": apiQuery.Get("content_type"),
	// })
	if apiQuery.Get("content_type") == "" {
		apiQuery.Set("content_type", contentTypes)
	}

	return apiQuery
}

func hasFilters(ctx context.Context, validatedQueryParams SearchURLParams) bool {
	if len(validatedQueryParams.Filter.Query) > 0 || len(validatedQueryParams.TopicFilter) > 0 {
		return true
	}

	return false
}

func createSearchAPIQuery(validatedQueryParams SearchURLParams) url.Values {
	return url.Values{
		"q":                []string{validatedQueryParams.Query},
		"population_types": []string{validatedQueryParams.PopulationTypeFilter},
		"dimensions":       []string{validatedQueryParams.DimensionsFilter},
		"content_type":     validatedQueryParams.Filter.Query,
		"sort":             []string{validatedQueryParams.Sort.Query},
		"limit":            []string{strconv.Itoa(validatedQueryParams.Limit)},
		"offset":           []string{strconv.Itoa(validatedQueryParams.Offset)},
		"topics":           []string{validatedQueryParams.TopicFilter},
	}
}

func createSearchControllerQuery(validatedQueryParams SearchURLParams) url.Values {
	return url.Values{
		"q":                []string{validatedQueryParams.Query},
		"population_types": []string{validatedQueryParams.PopulationTypeFilter},
		"dimensions":       []string{validatedQueryParams.DimensionsFilter},
		"filter":           validatedQueryParams.Filter.Query,
		"sort":             []string{validatedQueryParams.Sort.Query},
		"limit":            []string{strconv.Itoa(validatedQueryParams.Limit)},
		"page":             []string{strconv.Itoa(validatedQueryParams.CurrentPage)},
	}
}
