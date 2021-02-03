package data

const (
	relevance   = "relevance"
	releaseDate = "release_date"
	title       = "title"
)

// Sort represents information of a particular sort option
type Sort struct {
	Query           string `json:"query,omitempty"`
	LocaliseKeyName string `json:"localise_key"`
}

// SortOptions represent the list of all search sort options
var SortOptions = []Sort{Relevance, ReleaseDate, Title}

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
