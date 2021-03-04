package data

import (
	"context"
	"net/url"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/log.go/log"
)

// Sort represents information of a particular sort option
type Sort struct {
	Query           string `json:"query,omitempty"`
	LocaliseKeyName string `json:"localise_key"`
}

// SortOptions represent the list of all search sort options
var SortOptions = []Sort{Relevance, ReleaseDate, Title}

var sortOptions = map[string]bool{
	Relevance.Query:   true,
	ReleaseDate.Query: true,
	Title.Query:       true,
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

// ReviewSort retrieves sort from query and checks if it is one of the sort options
func ReviewSort(ctx context.Context, cfg *config.Config, query url.Values) {

	sortQuery := query.Get("sort")

	if !sortOptions[sortQuery] {
		log.Event(ctx, "sort chosen not available in sort options - default to sort "+cfg.DefaultSort, log.INFO)
		query.Set("sort", cfg.DefaultSort)
	}
}
