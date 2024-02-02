package data

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

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
	AfterDate            Date
	BeforeDate           Date
	TopicFilter          string
	LatestRelease        bool
	Sort                 Sort
	Limit                int
	CurrentPage          int
	Offset               int
}

var (
	dayValidator   = getIntValidator(1, 31)
	monthValidator = getIntValidator(1, 12)
	yearValidator  = getIntValidator(1900, 2150)
)

type intValidator func(valueAsString string) (int, error)

// getIntValidator returns an IntValidator object using the min and max values provided
func getIntValidator(minValue, maxValue int) intValidator {
	return func(valueAsString string) (int, error) {
		value, err := strconv.Atoi(valueAsString)
		if err != nil {
			return 0, fmt.Errorf("value contains non numeric characters")
		}
		if value < minValue {
			return 0, fmt.Errorf("value is below the minimum value (%d)", minValue)
		}
		if value > maxValue {
			return 0, fmt.Errorf("value is above the maximum value (%d)", maxValue)
		}

		return value, nil
	}
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
func ReviewDataAggregationQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	fromDate, toDate, err := GetDates(ctx, urlQuery)
	if err != nil {
		log.Error(ctx, "invalid dates set", err)
		return validatedQueryParams, err
	}
	validatedQueryParams.AfterDate = fromDate
	validatedQueryParams.BeforeDate = toDate

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

// ReviewDataAggregationQueryWithParams ensures that all search parameter values given by the user are reviewed
func ReviewDataAggregationQueryWithParams(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	fromDate, toDate, err := GetDates(ctx, urlQuery)
	if err != nil {
		log.Error(ctx, "invalid dates set", err)
		return validatedQueryParams, err
	}
	validatedQueryParams.AfterDate = fromDate
	validatedQueryParams.BeforeDate = toDate

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
	// TODO pass datatopiccache instead
	topicFilterErr := reviewTopicFiltersForDataAggregation(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
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
	}

	if apiQuery.Get("content_type") == "" {
		apiQuery.Set("content_type", contentTypes)
	}

	return apiQuery
}

// GetDates finds the date from and date to parameters
func GetDates(ctx context.Context, params url.Values) (startDate, endDate Date, err error) {
	var (
		startTime, endTime time.Time
	)

	const (
		DayBefore   = "before-day"
		DayAfter    = "after-day"
		MonthBefore = "before-month"
		MonthAfter  = "after-month"
		YearBefore  = "before-year"
		YearAfter   = "after-year"
		DateFrom    = "fromDate"
		DateTo      = "toDate"
	)

	yearAfterString, monthAfterString, dayAfterString := params.Get(YearAfter), params.Get(MonthAfter), params.Get(DayAfter)
	yearBeforeString, monthBeforeString, dayBeforeString := params.Get(YearBefore), params.Get(MonthBefore), params.Get(DayBefore)
	logData := log.Data{
		"year_after": yearAfterString, "month_after": monthAfterString, "day_after": dayAfterString,
		"year_before": yearBeforeString, "month_before": monthBeforeString, "day_before": DayBefore,
	}

	startTime, err = getValidTimestamp(yearAfterString, monthAfterString, dayAfterString)
	if err != nil {
		log.Warn(ctx, "invalid date, startDate", log.FormatErrors([]error{err}), logData)
		return Date{}, Date{}, err
	}

	startDate = DateFromTime(startTime)

	endTime, err = getValidTimestamp(yearBeforeString, monthBeforeString, dayBeforeString)
	if err != nil {
		log.Warn(ctx, "invalid date, endDate", log.FormatErrors([]error{err}), logData)
		return Date{}, Date{}, err
	}

	endDate = DateFromTime(endTime)

	if !startTime.IsZero() && !endTime.IsZero() && startTime.After(endTime) {
		log.Warn(ctx, "invalid date range: start date after end date", log.Data{DateFrom: startDate, DateTo: endDate})
		return Date{}, Date{}, errors.New("invalid dates: start date after end date")
	}

	return startDate, endDate, nil
}

func getValidTimestamp(year, month, day string) (time.Time, error) {
	if year == "" || month == "" || day == "" {
		return time.Time{}, nil
	}

	y, err := yearValidator(year)
	if err != nil {
		log.Error(context.TODO(), "Invalid parameter", err)
	}

	m, err := monthValidator(month)
	if err != nil {
		log.Error(context.TODO(), "Invalid parameter", err)
	}

	d, err := dayValidator(day)
	if err != nil {
		log.Error(context.TODO(), "Invalid parameter", err)
	}

	timestamp := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)

	// Check the day is valid for the month in the year, e.g. day 30 cannot be in month 2 (February)
	_, mo, _ := timestamp.Date()
	if mo != time.Month(m) {
		log.Error(context.TODO(), "Invalid parameter", err)
	}

	return timestamp, nil
}

func hasFilters(validatedQueryParams SearchURLParams) bool {
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
		"fromDate":         []string{validatedQueryParams.AfterDate.String()},
		"toDate":           []string{validatedQueryParams.BeforeDate.String()},
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
