package data

import (
	"context"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// SearchURLParams is a struct which contains all information of search url parameters and values
type SearchURLParams struct {
	Query       string
	Filter      Filter
	TopicFilter string
	Sort        Sort
	Limit       int
	CurrentPage int
	Offset      int
}

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	err := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if err != nil {
		log.Error(ctx, "unable to review pagination", err)
		return validatedQueryParams, err
	}

	reviewSort(ctx, cfg, urlQuery, &validatedQueryParams)

	err = reviewFilters(ctx, urlQuery, &validatedQueryParams)
	if err != nil {
		log.Error(ctx, "unable to review filters", err)
		return validatedQueryParams, err
	}

	err = reviewTopicFilters(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
	if err != nil {
		log.Error(ctx, "unable to review topic filters", err)
		return validatedQueryParams, err
	}

	err = reviewQueryString(ctx, urlQuery)
	if err != nil {
		log.Info(ctx, "the query string did not pass review")
		return validatedQueryParams, err
	}

	return validatedQueryParams, nil
}

// GetSearchAPIQuery gets the query that needs to be passed to the search-api to get search results
func GetSearchAPIQuery(validatedQueryParams SearchURLParams) url.Values {
	apiQuery := createSearchAPIQuery(validatedQueryParams)

	// update content_type query (filters) with sub filters
	updateQueryWithAPIFilters(apiQuery)

	// TODO finish the process for querying topics once the API is up and running, currently apiQuery will not include topics and so does nothing
	// update topics query with subtopics
	updateQueryWithAPITopics(apiQuery)

	return apiQuery
}

func createSearchAPIQuery(validatedQueryParams SearchURLParams) url.Values {
	return url.Values{
		"q":            []string{validatedQueryParams.Query},
		"content_type": validatedQueryParams.Filter.Query,
		"sort":         []string{validatedQueryParams.Sort.Query},
		"limit":        []string{strconv.Itoa(validatedQueryParams.Limit)},
		"offset":       []string{strconv.Itoa(validatedQueryParams.Offset)},
		"topics":       []string{validatedQueryParams.TopicFilter},
	}
}

func createSearchControllerQuery(validatedQueryParams SearchURLParams) url.Values {
	return url.Values{
		"q":      []string{validatedQueryParams.Query},
		"filter": validatedQueryParams.Filter.Query,
		"sort":   []string{validatedQueryParams.Sort.Query},
		"limit":  []string{strconv.Itoa(validatedQueryParams.Limit)},
		"page":   []string{strconv.Itoa(validatedQueryParams.CurrentPage)},
	}
}
