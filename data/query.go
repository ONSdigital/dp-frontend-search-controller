package data

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
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

// LimitOptions contains all available limit parameter values
var LimitOptions = []int{
	10,
	25,
	50,
}

// UpdateQueryWithOffset - removes page key and adds offset key to query to be then passed to dp-search-query
func updateQueryWithOffset(ctx context.Context, page *PaginationQuery, query url.Values) url.Values {

	offset := page.getOffset()

	updateQuery := query
	updateQuery.Set("offset", strconv.Itoa(offset))
	updateQuery.Del("page")

	return updateQuery
}

func (page *PaginationQuery) getOffset() int {
	return (page.CurrentPage - 1) * page.Limit
}

// ReviewQuery ensures that all empty query fields are set to default
func ReviewQuery(ctx context.Context, cfg *config.Config, url *url.URL) (*url.URL, *PaginationQuery, error) {
	query := url.Query()

	paginationQuery, err := reviewPagination(ctx, cfg, query)
	if err != nil {
		return url, nil, err
	}

	url.RawQuery = query.Encode()

	reviewSort(ctx, cfg, query)

	return url, paginationQuery, nil
}

func reviewPagination(ctx context.Context, cfg *config.Config, query url.Values) (*PaginationQuery, error) {
	page := getPage(ctx, query)

	limit := getLimit(ctx, cfg, query)

	if ((limit*page - 1) + 1) > cfg.DefaultMaximumSearchResults {
		return nil, errors.New("invalid page value, exceeding the default maximum search results")
	}

	paginationQuery := &PaginationQuery{
		Limit:       limit,
		CurrentPage: page,
	}

	return paginationQuery, nil
}

func getPage(ctx context.Context, query url.Values) int {

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, "unable to convert search page to int - set to default "+DefaultPageStr, log.INFO)
		query.Set("page", DefaultPageStr)
		page = DefaultPage
	}

	if page < 1 {
		log.Event(ctx, "page number is less than default - default to page "+DefaultPageStr, log.INFO)
		query.Set("page", DefaultPageStr)
		page = DefaultPage
	}

	return page
}

func getLimit(ctx context.Context, cfg *config.Config, query url.Values) int {

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, "unable to convert search limit to int - set to default "+DefaultLimitStr, log.INFO)
		query.Set("limit", DefaultLimitStr)
		limit = DefaultLimit
	}

	if limit < cfg.DefaultLimit {
		limit = cfg.DefaultLimit
	}

	if limit > cfg.DefaultMaximumLimit {
		limit = cfg.DefaultMaximumLimit
	}

	return limit
}

func reviewSort(ctx context.Context, cfg *config.Config, query url.Values) {

	sortQuery := query.Get("sort")

	if !sortOptions[sortQuery] {
		log.Event(ctx, "sort chosen not available in sort options - default to sort "+DefaultSort, log.INFO)
		query.Set("sort", DefaultSort)
	}
}
