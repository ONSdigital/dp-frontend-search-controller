package data

import (
	"context"
	"net/url"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/log"
)

// Category represents all the search categories in search page
type Category struct {
	LocaliseKeyName string        `json:"localise_key"`
	Count           int           `json:"count"`
	ContentTypes    []ContentType `json:"content_types"`
}

// ContentType represents the type of the search results and the number of results for each type
type ContentType struct {
	LocaliseKeyName string   `json:"localise_key"`
	Count           int      `json:"count"`
	Type            string   `json:"type"`
	SubTypes        []string `bson:"sub_types" json:"sub_types"`
}

// Categories represent the list of all search categories
var Categories = []Category{Publication, Data, Other}

// Publication - search information on publication category
var Publication = Category{
	LocaliseKeyName: "Publication",
	ContentTypes:    []ContentType{Bulletin, Article, Compendium},
}

// Data - search information on data category
var Data = Category{
	LocaliseKeyName: "Data",
	ContentTypes:    []ContentType{TimeSeries, Datasets, UserRequestedData},
}

// Other - search information on other categories
var Other = Category{
	LocaliseKeyName: "Other",
	ContentTypes:    []ContentType{Methodology, CorporateInformation},
}

// Bulletin - Search information specific for statistical bulletins
var Bulletin = ContentType{
	LocaliseKeyName: "StatisticalBulletin",
	Type:            "bulletin",
	SubTypes:        []string{"bulletin"},
}

// Article - Search information specific for articles
var Article = ContentType{
	LocaliseKeyName: "Article",
	Type:            "article",
	SubTypes:        []string{"article", "article_download"},
}

// Compendium - Search information specific for compendium
var Compendium = ContentType{
	LocaliseKeyName: "Compendium",
	Type:            "compendia",
	SubTypes:        []string{"compendium_landing_page"},
}

// TimeSeries - Search information specific for time series
var TimeSeries = ContentType{
	LocaliseKeyName: "TimeSeries",
	Type:            "time_series",
	SubTypes:        []string{"timeseries"},
}

// Datasets - Search information specific for datasets
var Datasets = ContentType{
	LocaliseKeyName: "Datasets",
	Type:            "datasets",
	SubTypes:        []string{"dataset_landing_page", "reference_tables"},
}

// UserRequestedData - Search information specific for user requested data
var UserRequestedData = ContentType{
	LocaliseKeyName: "UserRequestedData",
	Type:            "user_requested_data",
	SubTypes:        []string{"static_adhoc"},
}

// Methodology - Search information specific for methodologies
var Methodology = ContentType{
	LocaliseKeyName: "Methodology",
	Type:            "methodology",
	SubTypes:        []string{"static_methodology", "static_methodology_download", "static_qmi"},
}

// CorporateInformation - Search information specific for corporate information
var CorporateInformation = ContentType{
	LocaliseKeyName: "CorporateInformation",
	Type:            "corporate_information",
	SubTypes:        []string{"static_foi", "static_page", "static_landing_page", "static_article"},
}

// GetAllCategories returns all the categories and its content types where all the count is set to zero
func GetAllCategories() []Category {
	return setCountZero(Categories)
}

func setCountZero(categories []Category) []Category {
	for i, category := range categories {
		categories[i].Count = 0

		for j := range category.ContentTypes {
			categories[i].ContentTypes[j].Count = 0
		}

	}

	return categories
}

// GetSearchAPIQuery gets the query that needs to be passed to the search-api to get search results
func GetSearchAPIQuery(ctx context.Context, cfg *config.Config, page *PaginationQuery, query url.Values) (apiQuery url.Values, err error) {
	apiQuery, err = url.ParseQuery(query.Encode())
	if err != nil {
		log.Event(ctx, "failed to parse copy of query for search-api", log.Error(err), log.ERROR)
		return nil, err
	}

	// update query with offset and remove page query
	err = updateQueryWithOffset(ctx, cfg, page, apiQuery)
	if err != nil {
		log.Event(ctx, "failed to update query with offset", log.Error(err), log.ERROR)
		return nil, err
	}

	// update query with content_type which equals to sub filters and remove filter query
	err = updateQueryWithAPIFilters(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "failed to update query with api filters", log.Error(err), log.ERROR)
		return nil, err
	}

	return apiQuery, nil
}

// updateQueryWithAPIFilters retrieves and adds all available sub filters which is related to the search filter given by the user
func updateQueryWithAPIFilters(ctx context.Context, apiQuery url.Values) (err error) {
	filters := apiQuery["filter"]

	if len(filters) > 0 {
		subFilters, err := getSubFilters(filters)
		if err != nil {
			log.Event(ctx, "failed to get sub filters to query", log.Error(err), log.ERROR)
			return err
		}

		apiQuery.Del("filter")
		apiQuery.Set("content_type", strings.Join(subFilters, ","))
	}

	return nil
}

// getSubFilters gets all available sub filters which is related to the search filter given by the user
func getSubFilters(filters []string) ([]string, error) {
	var subFilters = make([]string, 0)

	for _, fType := range filters {
		found := false

	categoryLoop:
		for _, category := range Categories {
			for _, contentType := range category.ContentTypes {
				if fType == contentType.Type {
					found = true
					subFilters = append(subFilters, contentType.SubTypes...)
					break categoryLoop
				}
			}
		}

		if !found {
			return nil, errs.ErrFilterNotFound
		}

	}

	return subFilters, nil
}
