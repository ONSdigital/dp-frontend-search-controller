package data

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dis-design-system-go/model"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
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

	// Log and error if expected result is too large
	if (page * limit) > cfg.DefaultMaximumSearchResults {
		err := errs.ErrPageExceedsTotalPages
		log.Info(ctx, "requested page exceeds maximum search results", log.Data{
			"page":       page,
			"limit":      limit,
			"maxResults": cfg.DefaultMaximumSearchResults,
		})
		return err
	}

	offset, err := getOffset(ctx, cfg, page, limit)
	if err != nil {
		log.Error(ctx, "unable to get offset", err)
		return err
	}
	validatedQueryParams.Offset = offset

	return nil
}

func getLimitFromURLQuery(ctx context.Context, cfg *config.Config, query url.Values) int {
	limitParam := query.Get("limit")
	if limitParam == "" {
		return cfg.DefaultLimit
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		log.Info(ctx, fmt.Sprintf("unable to convert search limit to int - set to default %d", cfg.DefaultLimit))
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
	pageParam := query.Get("page")
	if pageParam == "" {
		return cfg.DefaultPage
	}

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		log.Info(ctx, "unable to convert search page to int - set to default", log.Data{
			"default": cfg.DefaultPage,
			"page":    page,
		})
		query.Set("page", strconv.Itoa(cfg.DefaultPage))
		return cfg.DefaultPage
	}

	if page < cfg.DefaultPage {
		log.Info(ctx, "page number is less than default - set to default", log.Data{
			"default": cfg.DefaultPage,
			"page":    page,
		})
		query.Set("page", strconv.Itoa(cfg.DefaultPage))
		return cfg.DefaultPage
	}

	return page
}

func getOffset(ctx context.Context, cfg *config.Config, page, limit int) (int, error) {
	offset := (page - 1) * limit

	// Log and default offset if it's negative
	if offset < 0 {
		log.Warn(ctx, "Offset less than 0 - set to default offset", log.Data{
			"page":          page,
			"limit":         limit,
			"calculated":    offset,
			"defaultOffset": cfg.DefaultOffset,
		})
		return cfg.DefaultOffset, nil
	}

	// Log and error if the offset is too large
	if (offset - limit) > cfg.DefaultMaximumSearchResults {
		err := errs.ErrInvalidPage
		log.Error(ctx, "Offset is too large - exceeds maximum search results", err, log.Data{
			"currentPage": page,
			"limit":       limit,
			"calculated":  offset,
			"maxResults":  cfg.DefaultMaximumSearchResults,
		})
		return cfg.DefaultOffset, err
	}

	return offset, nil
}

// GetTotalPages gets the total pages of the search results
func GetTotalPages(cfg *config.Config, limit, count int) int {
	if count > cfg.DefaultMaximumSearchResults {
		return cfg.DefaultMaximumSearchResults / limit
	}
	return (count + limit - 1) / limit
}

// GetPagesToDisplay gets all the pages available for the search results
func GetPagesToDisplay(cfg *config.Config, req http.Request, validatedQueryParams SearchURLParams, totalPages int) []model.PageToDisplay {
	pagesToDisplay := make([]model.PageToDisplay, 0)

	currentPage := validatedQueryParams.CurrentPage

	startPage := getStartPage(cfg, currentPage, totalPages)

	endPage := getEndPage(startPage, totalPages)

	controllerQuery := createSearchControllerQuery(validatedQueryParams)
	query := controllerQuery.Get("q")
	populationTypes := controllerQuery.Get("population_types")
	dimensions := controllerQuery.Get("dimensions")
	queryString := buildQueryString(query, populationTypes, dimensions)

	for i := startPage; i <= endPage; i++ {
		pagesToDisplay = append(pagesToDisplay, model.PageToDisplay{
			PageNumber: i,
			URL:        getPageURL(queryString, req, i, controllerQuery),
		})
	}

	return pagesToDisplay
}

// GetFirstAndLastPages gets the first and last pages
func GetFirstAndLastPages(req http.Request, validatedQueryParams SearchURLParams, totalPages int) []model.PageToDisplay {
	firstAndLastPages := make([]model.PageToDisplay, 0)

	controllerQuery := createSearchControllerQuery(validatedQueryParams)
	query := controllerQuery.Get("q")
	populationTypes := controllerQuery.Get("population_types")
	dimensions := controllerQuery.Get("dimensions")
	queryString := buildQueryString(query, populationTypes, dimensions)

	// add first and last
	firstAndLastPages = append(firstAndLastPages, model.PageToDisplay{
		PageNumber: 1,
		URL:        getPageURL(queryString, req, 1, controllerQuery),
	}, model.PageToDisplay{
		PageNumber: totalPages,
		URL:        getPageURL(queryString, req, totalPages, controllerQuery),
	})

	return firstAndLastPages
}

func getStartPage(cfg *config.Config, currentPage, totalPages int) int {
	pageOffset := getPageOffset()

	startPage := currentPage - pageOffset

	if (currentPage <= pageOffset) || (totalPages < noOfPagesToDisplay) {
		startPage = cfg.DefaultPage
	} else if (currentPage == totalPages-1) || (currentPage == totalPages) {
		startPage = totalPages - noOfPagesToDisplay + 1
	}

	return startPage
}

func getPageOffset() int {
	return int(math.Round((float64(noOfPagesToDisplay) - 1) / 2))
}

func getEndPage(startPage, totalPages int) int {
	endPage := startPage + noOfPagesToDisplay - 1

	if totalPages < noOfPagesToDisplay {
		endPage = totalPages
	}

	return endPage
}

func getPageURL(queryString string, req http.Request, page int, controllerQuery url.Values) (pageQueryString string) {
	controllerQuery.Del("q")
	controllerQuery.Del("page")
	controllerQuery.Del("population_types")
	controllerQuery.Del("dimensions")

	pageParam := "&page=" + strconv.Itoa(page)

	// This includes all the query parameters excluding search query and current page
	filterSortLimitParams := controllerQuery.Encode()
	if filterSortLimitParams != "" {
		filterSortLimitParams = "&" + filterSortLimitParams
	}

	// The pageURL is structured so that search query comes first and current page is mentioned last to make it more readable and logical
	pageQueryString = req.URL.Path + "?" + queryString + filterSortLimitParams + pageParam

	return pageQueryString
}

func buildQueryString(query, populationTypes, dimensions string) string {
	var u url.URL
	q := u.Query()
	q.Set("q", query)
	if populationTypes != "" {
		q.Set("population_types", populationTypes)
	}
	if dimensions != "" {
		q.Set("dimensions", dimensions)
	}
	return q.Encode()
}
