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

// updateQueryWithOffset - gets offset, add offset to query and removes page query to be then passed to dp-search-api
func updateQueryWithOffset(ctx context.Context, cfg *config.Config, pagination *PaginationQuery, query url.Values) error {

	offset, err := getOffset(ctx, cfg, pagination)
	if err != nil {
		log.Event(ctx, "unable to get offset", log.Error(err), log.ERROR)
		return err
	}

	query.Set("offset", strconv.Itoa(offset))
	query.Del("page")

	return nil
}

func getOffset(ctx context.Context, cfg *config.Config, pagination *PaginationQuery) (offset int, err error) {

	offset = (pagination.CurrentPage - 1) * pagination.Limit

	// when the offset is negative due to negative current page number or limit
	if offset < 0 {
		log.Event(ctx, "offset less than 0 - defaulted to offset "+strconv.Itoa(cfg.DefaultOffset), log.INFO)
		offset = cfg.DefaultOffset
	}

	// when the offset is too big due to big current page number and/or limit
	if (offset - pagination.Limit) > cfg.DefaultMaximumSearchResults {
		err = errs.ErrInvalidPage
		logData := log.Data{"current_page": pagination.CurrentPage, "limit": pagination.Limit}

		log.Event(ctx, "offset is too big as large page and/or limit given", log.Error(err), log.ERROR, logData)

		return cfg.DefaultOffset, err
	}

	return offset, nil
}

// ReviewPagination reviews page and limit values and returns paginationQuery containing these values
func ReviewPagination(ctx context.Context, cfg *config.Config, query url.Values) *PaginationQuery {
	page := getPage(ctx, cfg, query)

	limit := getLimit(ctx, cfg, query)

	paginationQuery := &PaginationQuery{
		Limit:       limit,
		CurrentPage: page,
	}

	return paginationQuery
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
