package data

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/v2/log"

	core "github.com/ONSdigital/dp-renderer/v2/model"
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
	NLPWeightingEnabled  bool
}

const (
	Limit       = "limit"
	Page        = "page"
	Offset      = "offset"
	SortName    = "sort"
	DayBefore   = "before-day"
	DayAfter    = "after-day"
	Before      = "before"
	MonthBefore = Before + "-month"
	After       = "after"
	MonthAfter  = After + "-month"
	YearBefore  = "before-year"
	YearAfter   = "after-year"
	Keywords    = "keywords"
	Query       = "query"
	DateFrom    = "fromDate"
	DateFromErr = DateFrom + "-error"
	DateTo      = "toDate"
	DateToErr   = DateTo + "-error"
	Type        = "release-type"
	Census      = "census"
	Highlight   = "highlight"

	PaginationErr           = "pagination-error"
	ContentTypeFilterErr    = "filter-error"
	TopicFilterErr          = "topic-error"
	PopulationTypeFilterErr = "population-error"
	DimensionsFilterErr     = "dimensions-error"
	QueryStringErr          = "query-string-error"
)

var (
	dayValidator   = getIntValidator(1, 31)
	monthValidator = getIntValidator(1, 12)
	yearValidator  = getIntValidator(1900, 2150)
)

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (sp SearchURLParams, validationErrs []core.ErrorItem) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	validationErrs = handleValidationError(ctx, paginationErr, "unable to review pagination for aggregation", PaginationErr, validationErrs)

	reviewSort(ctx, urlQuery, &validatedQueryParams, cfg.DefaultSort)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &sp)
	validationErrs = handleValidationError(ctx, contentTypeFilterError, "invalid content type filters set", ContentTypeFilterErr, validationErrs)

	topicFilterErr := reviewTopicFilters(ctx, urlQuery, &sp, censusTopicCache)
	validationErrs = handleValidationError(ctx, topicFilterErr, "invalid topic filters set", TopicFilterErr, validationErrs)

	populationTypeFilterErr := reviewPopulationTypeFilters(urlQuery, &sp)
	validationErrs = handleValidationError(ctx, populationTypeFilterErr, "invalid population types set", PopulationTypeFilterErr, validationErrs)

	dimensionsFilterErr := reviewDimensionsFilters(urlQuery, &sp)
	validationErrs = handleValidationError(ctx, dimensionsFilterErr, "invalid dimensions set", DimensionsFilterErr, validationErrs)

	queryStringErr := reviewQueryString(ctx, urlQuery)
	if !hasFilters(validatedQueryParams) {
		validationErrs = handleValidationError(ctx, queryStringErr, "the query string did not pass review", QueryStringErr, validationErrs)
	}

	return validatedQueryParams, validationErrs
}

func handleValidationError(ctx context.Context, err error, description, id string, validationErrs []core.ErrorItem) []core.ErrorItem {
	if err != nil {
		log.Error(ctx, description, err)
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: CapitalizeFirstLetter(err.Error()),
			},
			ID: id,
		})
	}
	return validationErrs
}

// ReviewDataAggregationQueryWithParams ensures that all search parameter values given by the user are reviewed
func ReviewDataAggregationQueryWithParams(ctx context.Context, cfg *config.Config, urlQuery url.Values) (sp SearchURLParams, validationErrs []core.ErrorItem) {
	sp.Query = urlQuery.Get("q")

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &sp)
	validationErrs = handleValidationError(ctx, paginationErr, "unable to review pagination for aggregation", PaginationErr, validationErrs)

	fromDate, vErrs := GetStartDate(urlQuery)
	if len(vErrs) > 0 {
		validationErrs = append(validationErrs, vErrs...)
	}
	sp.AfterDate = fromDate

	toDate, vErrs := GetEndDate(urlQuery)
	if len(vErrs) > 0 {
		validationErrs = append(validationErrs, vErrs...)
	}
	if fromDate.String() != "" && toDate.String() != "" {
		var err error
		toDate, err = ValidateDateRange(fromDate, toDate)
		if err != nil {
			validationErrs = append(validationErrs, core.ErrorItem{
				Description: core.Localisation{
					Text: CapitalizeFirstLetter(err.Error()),
				},
				ID:  DateToErr,
				URL: fmt.Sprintf("#%s", DateToErr),
			})
			toDate.fieldsetErrID = DateToErr
		}
	}
	sp.BeforeDate = toDate

	reviewSort(ctx, urlQuery, &sp, cfg.DefaultAggregationSort)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &sp)
	validationErrs = handleValidationError(ctx, contentTypeFilterError, "invalid content type filters set for aggregation", ContentTypeFilterErr, validationErrs)

	topicFilterErr := reviewTopicFiltersForDataAggregation(urlQuery, &sp)
	validationErrs = handleValidationError(ctx, topicFilterErr, "invalid topic filters set for aggregation", TopicFilterErr, validationErrs)

	populationTypeFilterErr := reviewPopulationTypeFilters(urlQuery, &sp)
	validationErrs = handleValidationError(ctx, populationTypeFilterErr, "invalid population types set for aggregation", PopulationTypeFilterErr, validationErrs)

	dimensionsFilterErr := reviewDimensionsFilters(urlQuery, &sp)
	validationErrs = handleValidationError(ctx, dimensionsFilterErr, "invalid dimensions set for aggregation", DimensionsFilterErr, validationErrs)

	return sp, validationErrs
}

// ReviewQuery ensures that all search parameter values given by the user are reviewed
func ReviewDatasetQuery(ctx context.Context, cfg *config.Config, urlQuery url.Values, censusTopicCache *cache.Topic) (SearchURLParams, error) {
	var validatedQueryParams SearchURLParams
	validatedQueryParams.Query = urlQuery.Get("q")

	paginationErr := reviewPagination(ctx, cfg, urlQuery, &validatedQueryParams)
	if paginationErr != nil {
		log.Error(ctx, "unable to review pagination for dataset", paginationErr)
		return validatedQueryParams, paginationErr
	}

	reviewSort(ctx, urlQuery, &validatedQueryParams, cfg.DefaultDatasetSort)

	contentTypeFilterError := reviewFilters(ctx, urlQuery, &validatedQueryParams)
	if contentTypeFilterError != nil {
		log.Error(ctx, "invalid content type filters set for dataset", contentTypeFilterError)
		return validatedQueryParams, contentTypeFilterError
	}
	topicFilterErr := reviewTopicFilters(ctx, urlQuery, &validatedQueryParams, censusTopicCache)
	if topicFilterErr != nil {
		log.Error(ctx, "invalid topic filters set for dataset", topicFilterErr)
		return validatedQueryParams, topicFilterErr
	}
	populationTypeFilterErr := reviewPopulationTypeFilters(urlQuery, &validatedQueryParams)
	if populationTypeFilterErr != nil {
		log.Error(ctx, "invalid population types set for dataset", populationTypeFilterErr)
		return validatedQueryParams, populationTypeFilterErr
	}
	dimensionsFilterErr := reviewDimensionsFilters(urlQuery, &validatedQueryParams)
	if dimensionsFilterErr != nil {
		log.Error(ctx, "invalid dimensions set for dataset", dimensionsFilterErr)
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
	} else {
		updateQueryWithAPIFilters(apiQuery)
	}

	return apiQuery
}

// GetStartDate returns the validated date from parameters
func GetStartDate(params url.Values) (startDate Date, validationErrs []core.ErrorItem) {
	var startTime time.Time

	startDate.fieldsetErrID = DateFromErr
	startDate.fieldsetStr = After

	yearAfterString, monthAfterString, dayAfterString := params.Get(YearAfter), params.Get(MonthAfter), params.Get(DayAfter)
	startDate.ds = dayAfterString
	startDate.ms = monthAfterString
	startDate.ys = yearAfterString

	if (monthAfterString != "" || dayAfterString != "") && yearAfterString == "" {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: "Enter the released after year",
			},
			ID:  DateFromErr,
			URL: fmt.Sprintf("#%s", DateFromErr),
		})
		startDate.hasYearValidationErr = true
		return startDate, validationErrs
	}

	var assumedDay, assumedMonth bool
	if yearAfterString != "" && monthAfterString == "" {
		monthAfterString = "1"
		assumedMonth = true
	}

	if yearAfterString != "" && dayAfterString == "" {
		dayAfterString = "1"
		assumedDay = true
	}

	startTime, validationErrs = getValidTimestamp(yearAfterString, monthAfterString, dayAfterString, &startDate)
	if len(validationErrs) > 0 {
		return startDate, validationErrs
	}

	startDate = DateFromTime(startTime)
	startDate.assumedDay = assumedDay
	startDate.assumedMonth = assumedMonth

	return startDate, nil
}

func GetEndDate(params url.Values) (endDate Date, validationErrs []core.ErrorItem) {
	var endTime time.Time

	endDate.fieldsetErrID = DateToErr
	endDate.fieldsetStr = Before

	yearBeforeString, monthBeforeString, dayBeforeString := params.Get(YearBefore), params.Get(MonthBefore), params.Get(DayBefore)
	endDate.ds = dayBeforeString
	endDate.ms = monthBeforeString
	endDate.ys = yearBeforeString

	if (monthBeforeString != "" || dayBeforeString != "") && yearBeforeString == "" {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: "Enter the released before year",
			},
			ID:  DateToErr,
			URL: fmt.Sprintf("#%s", DateToErr),
		})
		endDate.hasYearValidationErr = true
		return endDate, validationErrs
	}

	var assumedDay, assumedMonth bool
	if yearBeforeString != "" && monthBeforeString == "" {
		monthBeforeString = "1"
		assumedMonth = true
	}

	if yearBeforeString != "" && dayBeforeString == "" {
		dayBeforeString = "1"
		assumedDay = true
	}

	endTime, validationErrs = getValidTimestamp(yearBeforeString, monthBeforeString, dayBeforeString, &endDate)
	if len(validationErrs) > 0 {
		return endDate, validationErrs
	}

	endDate = DateFromTime(endTime)
	endDate.assumedDay = assumedDay
	endDate.assumedMonth = assumedMonth

	return endDate, nil
}

// ValidateDateRange returns an error and 'to' date if the 'from' date is after than the 'to' date
func ValidateDateRange(from, to Date) (end Date, err error) {
	startDate, err := ParseDate(from.String())
	if err != nil {
		return Date{}, err
	}
	endDate, err := ParseDate(to.String())
	if err != nil {
		return Date{}, err
	}

	startTime, _ := getValidTimestamp(startDate.YearString(), startDate.MonthString(), startDate.DayString(), &Date{})
	endTime, _ := getValidTimestamp(endDate.YearString(), endDate.MonthString(), endDate.DayString(), &Date{})
	if startTime.After(endTime) {
		end = to
		end.hasYearValidationErr = true
		return end, fmt.Errorf("enter a released before year that is later than %s", startDate.YearString())
	}
	return to, nil
}

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

// getValidTimestamp returns a valid timestamp or an error
func getValidTimestamp(year, month, day string, date *Date) (time.Time, []core.ErrorItem) {
	if year == "" || month == "" || day == "" {
		return time.Time{}, []core.ErrorItem{}
	}

	var validationErrs []core.ErrorItem

	d, err := dayValidator(day)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: fmt.Sprintf("%s for released %s day", CapitalizeFirstLetter(err.Error()), date.fieldsetStr),
			},
			ID:  date.fieldsetErrID,
			URL: fmt.Sprintf("#%s", date.fieldsetErrID),
		})
		date.hasDayValidationErr = true
	}

	m, err := monthValidator(month)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: fmt.Sprintf("%s for released %s month", CapitalizeFirstLetter(err.Error()), date.fieldsetStr),
			},
			ID:  date.fieldsetErrID,
			URL: fmt.Sprintf("#%s", date.fieldsetErrID),
		})
		date.hasMonthValidationErr = true
	}

	y, err := yearValidator(year)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: fmt.Sprintf("%s for released %s year", CapitalizeFirstLetter(err.Error()), date.fieldsetStr),
			},
			ID:  date.fieldsetErrID,
			URL: fmt.Sprintf("#%s", date.fieldsetErrID),
		})
		date.hasYearValidationErr = true
	}

	// Throw errors back to user before further validation
	if len(validationErrs) > 0 {
		return time.Time{}, validationErrs
	}

	timestamp := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)

	// Check the day is valid for the month in the year, e.g. day 30 cannot be in month 2 (February)
	_, mo, _ := timestamp.Date()
	if mo != time.Month(m) {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				LocaleKey: "ValidationInvalidDate",
				Plural:    1,
			},
			ID:  date.fieldsetErrID,
			URL: fmt.Sprintf("#%s", date.fieldsetErrID),
		})
		date.hasDayValidationErr = true
		date.hasMonthValidationErr = true
		date.hasYearValidationErr = true
	}
	return timestamp, validationErrs
}

func hasFilters(validatedQueryParams SearchURLParams) bool {
	if len(validatedQueryParams.Filter.Query) > 0 || validatedQueryParams.TopicFilter != "" {
		return true
	}

	return false
}

// CapitalizeFirstLetter is a helper function that transforms the first letter of a string to uppercase
func CapitalizeFirstLetter(input string) string {
	switch {
	case input == "":
		return input
	case len(input) == 1:
		return strings.ToUpper(input)
	default:
		return strings.ToUpper(input[:1]) + input[1:]
	}
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
		"nlp_weighting":    []string{strconv.FormatBool(validatedQueryParams.NLPWeightingEnabled)},
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
