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

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if paginationErr != nil {
		log.Error(ctx, "unable to review pagination", paginationErr)
		return validatedQueryParams, paginationErr
	}

	reviewSort(ctx, cfg, urlQuery, &validatedQueryParams)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &validatedQueryParams)
	topicFilterErr := reviewTopicFilters(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
	if contentTypeFilterError != nil && topicFilterErr != nil {
		log.Error(ctx, "unable to review both content type and topic filters", apperrors.ErrInvalidConentTypeAndTopicFilters)
		return validatedQueryParams, apperrors.ErrInvalidConentTypeAndTopicFilters
	}

	queryStringErr := reviewQueryString(ctx, urlQuery)
	if queryStringErr != nil && errors.Is(queryStringErr, apperrors.ErrInvalidQueryCharLengthString) {
		log.Info(ctx, "the query string did not pass review")
		return validatedQueryParams, queryStringErr
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
