package data

import (
	"context"
	"net/url"
	"strconv"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/log"
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

	reviewSort(ctx, cfg, query)

	url.RawQuery = query.Encode()

	return url, paginationQuery, nil
}

func reviewPagination(ctx context.Context, cfg *config.Config, query url.Values) (*PaginationQuery, error) {
	page := getPage(ctx, cfg, query)

	limit := getLimit(ctx, cfg, query)

	if ((limit*page - 1) + 1) > cfg.DefaultMaximumSearchResults {
		return nil, errs.ErrInvalidPage
	}

	paginationQuery := &PaginationQuery{
		Limit:       limit,
		CurrentPage: page,
	}

	return paginationQuery, nil
}

func getPage(ctx context.Context, cfg *config.Config, query url.Values) int {
	defaultPageStr := strconv.Itoa(cfg.DefaultPage)

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, "unable to convert search page to int - set to default "+defaultPageStr, log.INFO)
		query.Set("page", defaultPageStr)
		page = cfg.DefaultPage
	}

	if page < 1 {
		log.Event(ctx, "page number is less than default - default to page "+defaultPageStr, log.INFO)
		query.Set("page", defaultPageStr)
		page = cfg.DefaultPage
	}

	return page
}

func getLimit(ctx context.Context, cfg *config.Config, query url.Values) int {
	defaultLimitStr := strconv.Itoa(cfg.DefaultLimit)

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, "unable to convert search limit to int - set to default "+defaultLimitStr, log.INFO)
		query.Set("limit", defaultLimitStr)
		limit = cfg.DefaultLimit
	}

	if limit < cfg.DefaultLimit {
		query.Set("limit", strconv.Itoa(cfg.DefaultLimit))
		limit = cfg.DefaultLimit
	}

	if limit > cfg.DefaultMaximumLimit {
		query.Set("limit", strconv.Itoa(cfg.DefaultMaximumLimit))
		limit = cfg.DefaultMaximumLimit
	}

	return limit
}

func reviewSort(ctx context.Context, cfg *config.Config, query url.Values) {

	sortQuery := query.Get("sort")

	if !sortOptions[sortQuery] {
		log.Event(ctx, "sort chosen not available in sort options - default to sort "+cfg.DefaultSort, log.INFO)
		query.Set("sort", cfg.DefaultSort)
	}
}
