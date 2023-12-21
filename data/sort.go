package data

import (
	"context"
	"net/url"

	"github.com/ONSdigital/log.go/v2/log"
)

// Sort represents information of a particular sort option
type Sort struct {
	Query           string `json:"query"`
	LocaliseKeyName string `json:"localise_key"`
}

var (
	// SortOptions represent the list of all search sort options
	SortOptions        = []Sort{Relevance, ReleaseDate, Title}
	DatasetSortOptions = []Sort{ReleaseDate, Title}

	// Relevance - informing on sorting based on relevance
	Relevance = Sort{
		Query:           "relevance",
		LocaliseKeyName: "Relevance",
	}

	// ReleaseDate - informing on sorting based on release date
	ReleaseDate = Sort{
		Query:           "release_date",
		LocaliseKeyName: "ReleaseDate",
	}

	// Title - informing on sorting based on title
	Title = Sort{
		Query:           "title",
		LocaliseKeyName: "Title",
	}

	// sortOptions contains all the possible sort available on the search page
	sortOptions = map[string]Sort{
		Relevance.Query:   Relevance,
		ReleaseDate.Query: ReleaseDate,
		Title.Query:       Title,
	}
)

// reviewSort retrieves sort from query and checks if it is one of the sort options
func reviewSort(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams, defaultSort string) {
	sortQuery := urlQuery.Get("sort")

	sort, found := sortOptions[sortQuery]

	if !found {
		log.Warn(ctx, "sort chosen not available in sort options - default to sort "+defaultSort)
		sort.Query = defaultSort
		sort.LocaliseKeyName = sortOptions[defaultSort].LocaliseKeyName
	}

	validatedQueryParams.Sort.Query = sort.Query
	validatedQueryParams.Sort.LocaliseKeyName = sort.LocaliseKeyName
}
