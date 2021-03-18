package data

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-frontend-models/model"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/log"
)

const noOfPagesToDisplay = 5

// LimitOptions contains all available limit parameter values
var LimitOptions = []int{
	10,
	25,
	50,
}

// reviewPagination reviews page and limit values and sets limit and page values to queryParams
func reviewPagination(ctx context.Context, cfg *config.Config, urlQuery url.Values, validatedQueryParams *SearchURLParams) error {
	limit := getLimitFromURLQuery(ctx, cfg, urlQuery)
	validatedQueryParams.Limit = limit

	page := getPageFromURLQuery(ctx, cfg, urlQuery)
	validatedQueryParams.CurrentPage = page

	offset, err := getOffset(ctx, cfg, page, limit)
	if err != nil {
		log.Event(ctx, "unable to get offset", log.Error(err), log.ERROR)
		return err
	}
	validatedQueryParams.Offset = offset

	return nil
}

func getLimitFromURLQuery(ctx context.Context, cfg *config.Config, query url.Values) int {
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, fmt.Sprintf("unable to convert search limit to int - set to default %d", cfg.DefaultLimit), log.INFO)
		query.Set("limit", strconv.Itoa(cfg.DefaultLimit))
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

func getPageFromURLQuery(ctx context.Context, cfg *config.Config, query url.Values) int {
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, fmt.Sprintf("unable to convert search page to int - set to default %d", cfg.DefaultPage), log.INFO)
		query.Set("page", strconv.Itoa(cfg.DefaultPage))
		page = cfg.DefaultPage
	}

	if page < 1 {
		log.Event(ctx, fmt.Sprintf("page number is less than default - default to page %d", cfg.DefaultPage), log.INFO)
		query.Set("page", strconv.Itoa(cfg.DefaultPage))
		page = cfg.DefaultPage
	}

	return page
}

func getOffset(ctx context.Context, cfg *config.Config, page int, limit int) (offset int, err error) {

	offset = (page - 1) * limit

	// when the offset is negative due to negative current page number or limit
	if offset < 0 {
		log.Event(ctx, fmt.Sprintf("offset less than 0 - defaulted to offset %d", cfg.DefaultOffset), log.WARN)
		offset = cfg.DefaultOffset
	}

	// when the offset is too big due to big current page number and/or limit
	if (offset - limit) > cfg.DefaultMaximumSearchResults {
		err = errs.ErrInvalidPage
		logData := log.Data{"current_page": page, "limit": limit}

		log.Event(ctx, "offset is too big as large page and/or limit given", log.Error(err), log.ERROR, logData)

		return cfg.DefaultOffset, err
	}

	return offset, nil
}

// GetTotalPages gets the total pages of the search results
func GetTotalPages(limit int, count int) int {
	return (count + limit - 1) / limit
}

// GetPagesToDisplay gets all the pages available for the search results
func GetPagesToDisplay(cfg *config.Config, validatedQueryParams SearchURLParams, totalPages int) []model.PageToDisplay {
	var pagesToDisplay = make([]model.PageToDisplay, 0)

	currentPage := validatedQueryParams.CurrentPage

	startPage := getStartPage(cfg, currentPage, totalPages)

	endPage := getEndPage(startPage, totalPages)

	controllerQuery := createSearchControllerQuery(validatedQueryParams)
	query := controllerQuery.Get("q")

	for i := startPage; i <= endPage; i++ {
		pagesToDisplay = append(pagesToDisplay, model.PageToDisplay{
			PageNumber: i,
			URL:        getPageURL(query, i, controllerQuery),
		})
	}

	return pagesToDisplay
}

func getStartPage(cfg *config.Config, currentPage int, totalPages int) int {
	pageOffset := getPageOffset()

	startPage := currentPage - pageOffset

	if currentPage <= pageOffset {
		startPage = cfg.DefaultPage
	} else if (currentPage == totalPages-1) || (currentPage == totalPages) {
		startPage = totalPages - noOfPagesToDisplay + 1
	}

	return startPage
}

func getPageOffset() int {
	return int(math.Round((float64(noOfPagesToDisplay) - 1) / 2))
}

func getEndPage(startPage int, totalPages int) int {
	endPage := startPage + noOfPagesToDisplay - 1

	if totalPages < noOfPagesToDisplay {
		endPage = totalPages
	}

	return endPage
}

func getPageURL(query string, page int, controllerQuery url.Values) (pageURL string) {
	controllerQuery.Del("q")
	controllerQuery.Del("page")

	queryParam := "q=" + query
	pageParam := "&page=" + strconv.Itoa(page)

	// This includes all the query parameters excluding search query and current page
	filterSortLimitParams := controllerQuery.Encode()
	if filterSortLimitParams != "" {
		filterSortLimitParams = "&" + filterSortLimitParams
	}

	// The pageURL is structured so that search query comes first and current page is mentioned last to make it more readable and logical
	pageURL = "/search?" + queryParam + filterSortLimitParams + pageParam

	return pageURL
}
