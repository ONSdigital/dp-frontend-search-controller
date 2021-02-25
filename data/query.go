package data

import (
	"context"
	"net/url"
	"strconv"

	"github.com/ONSdigital/log.go/log"
)

const (
	// DefaultLimit - default values for limit query
	DefaultLimit = 10
	// DefaultLimitStr - default values for limit query in string format
	DefaultLimitStr = "10"
	// DefaultSort - default values for sort query
	DefaultSort = "relevance"
	// DefaultPage - default values for page query
	DefaultPage = 1
	// DefaultPageStr - default values for page query in string format
	DefaultPageStr = "1"
)

// PaginationQuery is a struct which contains validated pagination information
type PaginationQuery struct {
	Limit       int
	CurrentPage int
}

// GetLimitOptions returns all available limit options for search
func GetLimitOptions() []int {
	return []int{10, 25, 50}
}

// UpdateQueryWithOffset - removes page key and adds offset key to query to be then passed to dp-search-query
func updateQueryWithOffset(ctx context.Context, query url.Values) url.Values {

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, "unable to convert search page to int - set to default "+DefaultPageStr, log.INFO)
		page = DefaultPage
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, "unable to convert search limit to int - set to default "+DefaultLimitStr, log.INFO)
		limit = DefaultLimit
	}
	offset := strconv.Itoa((page - 1) * limit)
	updateQuery := query
	updateQuery.Set("offset", offset)
	updateQuery.Del("page")
	return updateQuery
}

// SetDefaultQueries ensures that all empty query fields are set to default
func SetDefaultQueries(ctx context.Context, url *url.URL) (*url.URL, *PaginationQuery) {
	var found bool
	query := url.Query()

	// Validating current page
	currentPage, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, "unable to convert search page to int - set to default "+DefaultPageStr, log.INFO)
		query.Set("page", DefaultPageStr)
		currentPage = DefaultPage
	} else {
		if currentPage < 1 {
			log.Event(ctx, "page number is less than default - default to page "+DefaultPageStr, log.INFO)
			query.Set("page", DefaultPageStr)
			currentPage = DefaultPage
		}
	}

	// Validating search limit
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, "unable to convert search limit to int - set to default "+DefaultLimitStr, log.INFO)
		query.Set("limit", DefaultLimitStr)
		limit = DefaultLimit
	} else {
		found = false
		limitOptions := GetLimitOptions()
		for _, limitOption := range limitOptions {
			if limitOption == limit {
				found = true
			}
		}
		if !found {
			log.Event(ctx, "limit chosen not available in limit options - default to limit "+DefaultLimitStr, log.INFO)
			query.Set("limit", DefaultLimitStr)
			limit = DefaultLimit
		}
	}

	// Validating sort query
	sortQuery := query.Get("sort")
	if sortQuery == "" {
		query.Set("sort", DefaultSort)
	} else {
		found = false
		for _, sort := range SortOptions {
			if sortQuery == sort.Query {
				found = true
			}
		}
		if !found {
			query.Set("sort", DefaultSort)
			log.Event(ctx, "sort chosen not available in sort options - default to sort "+DefaultSort, log.INFO)
		}
	}

	url.RawQuery = query.Encode()

	paginationQuery := &PaginationQuery{
		Limit:       limit,
		CurrentPage: currentPage,
	}
	return url, paginationQuery
}
