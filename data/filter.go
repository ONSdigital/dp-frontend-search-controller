package data

import (
	"context"
	"errors"
	"net/url"
	"strings"

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

var errFilterType = errors.New("invalid filter type given")

func setCountZero(categories []Category) []Category {
	for i, category := range categories {
		categories[i].Count = 0
		for j := range category.ContentTypes {
			categories[i].ContentTypes[j].Count = 0
		}
	}
	return categories
}

// GetAllCategories returns all the categories and its content types where all the count is set to zero
func GetAllCategories() []Category {
	return setCountZero(Categories)
}

// MapSubFilterTypes - adds sub filter types to filter query to be then passed to logic to retrieve search results
func MapSubFilterTypes(ctx context.Context, page *PaginationQuery, query url.Values) (apiQuery url.Values, err error) {
	apiQuery = updateQueryWithOffset(ctx, page, query)
	apiQuery, err = url.ParseQuery(apiQuery.Encode())
	if err != nil {
		log.Event(ctx, "failed to parse copy of query for mapping filter types", log.Error(err), log.ERROR)
		return nil, err
	}
	filters := apiQuery["filter"]
	if len(filters) > 0 {
		var newFilters = make([]string, 0)
		for _, fType := range filters {
			found := false
		categoryLoop:
			for _, category := range Categories {
				for _, contentType := range category.ContentTypes {
					if fType == contentType.Type {
						found = true
						newFilters = append(newFilters, contentType.SubTypes...)
						break categoryLoop
					}
				}
			}
			if !found {
				return nil, errFilterType
			}
		}
		apiQuery.Del("filter")
		apiQuery.Set("content_type", strings.Join(newFilters, ","))
	}
	return apiQuery, nil
}
