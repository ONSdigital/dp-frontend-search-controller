package data

import (
	"context"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// SearchURLParams is a struct which contains all information of search url parameters and values
type SearchURLParams struct {
	Query       string
	Filter      Filter
	Sort        Sort
	Limit       int
	CurrentPage int
	Offset      int
}

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values) (SearchURLParams, bool, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	err := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if err != nil {
		log.Error(ctx, "unable to review pagination", err)
		return validatedQueryParams, false, err
	}

	reviewSort(ctx, cfg, urlQuery, &validatedQueryParams)

	err = reviewFilters(ctx, urlQuery, &validatedQueryParams)
	if err != nil {
		log.Error(ctx, "unable to review filters", err)
		return validatedQueryParams, false, err
	}

	validationProblem := reviewQueryString(ctx, urlQuery)

	return validatedQueryParams, validationProblem, nil
}

// GetSearchAPIQuery gets the query that needs to be passed to the search-api to get search results
func GetSearchAPIQuery(validatedQueryParams SearchURLParams) url.Values {
	apiQuery := createSearchAPIQuery(validatedQueryParams)

	// update content_type query (filters) with sub filters
	updateQueryWithAPIFilters(apiQuery)

	return apiQuery
}

func createSearchAPIQuery(validatedQueryParams SearchURLParams) url.Values {
	return url.Values{
		"q":            []string{validatedQueryParams.Query},
		"content_type": validatedQueryParams.Filter.Query,
		"sort":         []string{validatedQueryParams.Sort.Query},
		"limit":        []string{strconv.Itoa(validatedQueryParams.Limit)},
		"offset":       []string{strconv.Itoa(validatedQueryParams.Offset)},
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
