package data

import (
	"context"
	"net/url"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/log"
)

// Sort represents information of a particular sort option
type Sort struct {
	Query           string `json:"query"`
	LocaliseKeyName string `json:"localise_key"`
}

// SortOptions represent the list of all search sort options
var SortOptions = []Sort{Relevance, ReleaseDate, Title}

var sortOptions = map[string]Sort{
	Relevance.Query:   Relevance,
	ReleaseDate.Query: ReleaseDate,
	Title.Query:       Title,
}

// Relevance - informing on sorting based on relevance
var Relevance = Sort{
	Query:           "relevance",
	LocaliseKeyName: "Relevance",
}

// ReleaseDate - informing on sorting based on release date
var ReleaseDate = Sort{
	Query:           "release_date",
	LocaliseKeyName: "ReleaseDate",
}

// Title - informing on sorting based on title
var Title = Sort{
	Query:           "title",
	LocaliseKeyName: "Title",
}

// reviewSort retrieves sort from query and checks if it is one of the sort options
func reviewSort(ctx context.Context, cfg *config.Config, urlQuery url.Values, validatedQueryParams *SearchURLParams) {

	sortQuery := urlQuery.Get("sort")

	sort, found := sortOptions[sortQuery]

	if found {
		validatedQueryParams.Sort.Query = sort.Query
		validatedQueryParams.Sort.LocaliseKeyName = sort.LocaliseKeyName
	} else {
		log.Event(ctx, "sort chosen not available in sort options - default to sort "+cfg.DefaultSort, log.INFO)
		validatedQueryParams.Sort.Query = cfg.DefaultSort
		validatedQueryParams.Sort.LocaliseKeyName = sortOptions[cfg.DefaultSort].LocaliseKeyName
	}
}
