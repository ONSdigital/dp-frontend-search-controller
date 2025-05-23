package data

import (
	"context"
	"net/url"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// Filter represents information of filters selected by user
type Filter struct {
	Query           []string `json:"query,omitempty"`
	LocaliseKeyName []string `json:"localise_key,omitempty"`
}

// Category represents all the search categories in search page
type Category struct {
	LocaliseKeyName  string        `json:"localise_key"`
	Count            int           `json:"count"`
	ContentTypes     []ContentType `json:"content_types"`
	HideTypesInWebUI bool          `json:"hide_types_in_web_ui"`
	ContentTypeKey   string        `json:"content_type_key"`
}

// ContentType represents the type of the search results and the number of results for each type
type ContentType struct {
	LocaliseKeyName string   `json:"localise_key"`
	Count           int      `json:"count"`
	Group           string   `json:"group"`
	Types           []string `json:"types"`
	ShowInWebUI     bool     `json:"show_in_web_ui"`
}

var defaultContentTypes = "article," +
	"article_download," +
	"bulletin," +
	"compendium_landing_page," +
	"dataset_landing_page," +
	"product_page," +
	"static_adhoc," +
	"static_article," +
	"static_foi," +
	"static_landing_page," +
	"static_methodology," +
	"static_methodology_download," +
	"static_page," +
	"static_qmi," +
	"statistical_article," +
	"timeseries," +
	"timeseries_dataset"

var (
	// Categories represent the list of all search categories
	Categories = []Category{Publication, Data, Other}

	// Publication - search information on publication category
	Publication = Category{
		LocaliseKeyName:  "Publication",
		ContentTypes:     []ContentType{Bulletin, Article, Compendium, StatisticalArticle},
		HideTypesInWebUI: true,
		ContentTypeKey:   "publications",
	}

	// Data - search information on data category
	Data = Category{
		LocaliseKeyName: "Data",
		ContentTypes:    []ContentType{TimeSeries, Datasets, UserRequestedData},
	}

	// Other - search information on other categories
	Other = Category{
		LocaliseKeyName: "Other",
		ContentTypes:    []ContentType{Methodology, CorporateInformation, ProductPage},
	}

	// UnusedCategories - these categories are served from the Search API but we don't use them for filters.
	// If left in, they create noise in the logs about unexpected filters being returned.
	UnusedCategoryTypes = []string{
		"api_dataset_landing_page",
		"compendium_chapter",
		"compendium_data",
		"dataset",
		"home_page",
		"home_page_census",
		"reference_tables",
		"release",
		"taxonomy_landing_page",
		"visualisation",
	}

	// Bulletin - Search information specific for statistical bulletins
	Bulletin = ContentType{
		LocaliseKeyName: "StatisticalBulletin",
		Group:           "bulletin",
		Types:           []string{"bulletin"},
		ShowInWebUI:     true,
	}

	// Article - Search information specific for articles
	Article = ContentType{
		LocaliseKeyName: "Article",
		Group:           "article",
		Types:           []string{"article", "article_download"},
		ShowInWebUI:     true,
	}

	// Compendium - Search information specific for compendium
	Compendium = ContentType{
		LocaliseKeyName: "Compendium",
		Group:           "compendia",
		Types:           []string{"compendium_landing_page"},
		ShowInWebUI:     true,
	}

	// StatisticalArticle - search information specific for statistical articles
	StatisticalArticle = ContentType{
		LocaliseKeyName: "StatisticalArticle",
		Group:           "statistical_article",
		Types:           []string{"statistical_article"},
		ShowInWebUI:     true,
	}

	// TimeSeries - Search information specific for time series
	TimeSeries = ContentType{
		LocaliseKeyName: "TimeSeries",
		Group:           "time_series",
		Types:           []string{"timeseries"},
		ShowInWebUI:     true,
	}

	// Datasets - Search information specific for datasets
	Datasets = ContentType{
		LocaliseKeyName: "Datasets",
		Group:           "datasets",
		Types:           []string{"dataset_landing_page", "timeseries_dataset"},
		ShowInWebUI:     true,
	}

	DatasetLandingPage = ContentType{
		LocaliseKeyName: "Datasets",
		Group:           "dataset_landing_page",
		Types:           []string{"dataset_landing_page"},
		ShowInWebUI:     true,
	}

	// UserRequestedData - Search information specific for user requested data
	UserRequestedData = ContentType{
		LocaliseKeyName: "UserRequestedData",
		Group:           "user_requested_data",
		Types:           []string{"static_adhoc"},
		ShowInWebUI:     true,
	}

	// Methodology - Search information specific for methodologies
	Methodology = ContentType{
		LocaliseKeyName: "Methodology",
		Group:           "methodology",
		Types:           []string{"static_methodology", "static_methodology_download", "static_qmi"},
		ShowInWebUI:     true,
	}

	// CorporateInformation - Search information specific for corporate information
	CorporateInformation = ContentType{
		LocaliseKeyName: "CorporateInformation",
		Group:           "corporate_information",
		Types:           []string{"static_foi", "static_page", "static_landing_page", "static_article"},
		ShowInWebUI:     true,
	}

	// ProductPage - Search information specific for product pages
	ProductPage = ContentType{
		LocaliseKeyName: "ProductPage",
		Group:           "product_page",
		Types:           []string{"product_page"},
		ShowInWebUI:     false,
	}

	PublicationGroup = ContentType{
		Group:       Publication.ContentTypeKey,
		Types:       []string{"bulletin", "article", "article_download", "compendium_landing_page", "statistical_article"},
		ShowInWebUI: false,
	}

	// filterOptions contains all the possible filter available on the search page
	filterOptions = map[string]ContentType{
		Article.Group:              Article,
		Bulletin.Group:             Bulletin,
		Compendium.Group:           Compendium,
		CorporateInformation.Group: CorporateInformation,
		Datasets.Group:             Datasets,
		DatasetLandingPage.Group:   DatasetLandingPage,
		Methodology.Group:          Methodology,
		ProductPage.Group:          ProductPage,
		StatisticalArticle.Group:   StatisticalArticle,
		TimeSeries.Group:           TimeSeries,
		UserRequestedData.Group:    UserRequestedData,
		PublicationGroup.Group:     PublicationGroup,
	}
)

// reviewFilter retrieves filters from query, checks if they are one of the filter options, and updates validatedQueryParams
func reviewFilters(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams) (err error) {
	filtersQuery := urlQuery["filter"]

	for _, filterQuery := range filtersQuery {
		filterQuery = strings.ToLower(filterQuery)

		if filterQuery == "" {
			continue
		}

		filter, found := filterOptions[filterQuery]

		if !found {
			err = errs.ErrContentTypeNotFound
			logData := log.Data{"requested_filter": filter}
			log.Info(ctx, "failed to find filter", logData)

			break
		}

		validatedQueryParams.Filter.Query = append(validatedQueryParams.Filter.Query, filter.Group)
		validatedQueryParams.Filter.LocaliseKeyName = append(validatedQueryParams.Filter.LocaliseKeyName, filter.LocaliseKeyName)
	}

	return err
}

// GetCategories returns all the categories and its content types where all the count is set to zero
func GetCategories() []Category {
	var categories []Category
	categories = append(categories, Categories...)

	// To get a different reference of ContentType - deep copy
	for i, category := range categories {
		categories[i].ContentTypes = []ContentType{}
		categories[i].ContentTypes = append(categories[i].ContentTypes, Categories[i].ContentTypes...)

		// To get a different reference of SubTypes - deep copy
		for j := range category.ContentTypes {
			categories[i].ContentTypes[j].Types = []string{}
			categories[i].ContentTypes[j].Types = append(categories[i].ContentTypes[j].Types, Categories[i].ContentTypes[j].Types...)
		}
	}

	return categories
}

// updateQueryWithAPIFilters retrieves and adds all available sub filters which is related to the search filter given by the user
func updateQueryWithAPIFilters(apiQuery url.Values) {
	filters := apiQuery["content_type"]

	if len(filters) > 0 {
		subFilters := getSubFilters(filters)

		apiQuery.Set("content_type", strings.Join(subFilters, ","))
	} else {
		apiQuery.Set("content_type", defaultContentTypes)
	}
}

// getSubFilters gets all available sub filters which is related to the search filter given by the user
func getSubFilters(filters []string) []string {
	subFilters := make([]string, 0)

	for _, filter := range filters {
		subFilter := filterOptions[filter]
		subFilters = append(subFilters, subFilter.Types...)
	}

	return subFilters
}

// GetGroupLocaliseKey gets the localise key of the group type of the search result to be displayed
func GetGroupLocaliseKey(resultType string) string {
	for _, filterOption := range filterOptions {
		for _, optionType := range filterOption.Types {
			if resultType == optionType {
				if !filterOption.ShowInWebUI {
					continue
				}

				return filterOption.LocaliseKeyName
			}
		}
	}
	return ""
}

// IsCategoryUnused returns if a category is unused.
func IsCategoryUnused(categoryName string) bool {
	for _, val := range UnusedCategoryTypes {
		if val == categoryName {
			return true
		}
	}
	return false
}
