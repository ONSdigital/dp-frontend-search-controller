package model

import "github.com/ONSdigital/dp-renderer/model"

// Search is the model struct for the cookies preferences form
type SearchPage struct {
	model.Page
	Data       Search      `json:"data"`
	Department *Department `json:"department"`
}

// Search represents all search parameters and response data of the search
type Search struct {
	Query        string           `json:"query"`
	ErrorMessage string           `json:"error_message,omitempty"`
	Filter       []string         `json:"filter,omitempty"`
	Filters      []Filter         `json:"filters"`
	TopicFilters []Filter         `json:"topic_filters"`
	Sort         Sort             `json:"sort,omitempty"`
	Pagination   model.Pagination `json:"pagination,omitempty"`
	Response     Response         `json:"response"`
}

// Filter respresents all filter information needed by templates
type Filter struct {
	LocaliseKeyName string   `json:"localise_key_name,omitempty"`
	FilterKey       []string `json:"filter_key,omitempty"`
	IsChecked       bool     `json:"is_checked,omitempty"`
	NumberOfResults int      `json:"number_of_results,omitempty"`
	Types           []Filter `json:"types,omitempty"`
}

// Sort represents all the information of sorting related to the search page
type Sort struct {
	Query              string        `json:"query,omitempty"`
	LocaliseFilterKeys []string      `json:"filter_text,omitempty"`
	LocaliseSortKey    string        `json:"sort_text,omitempty"`
	Options            []SortOptions `json:"options,omitempty"`
}

// SortOptions represents all the information of different sorts available
type SortOptions struct {
	Query           string `json:"query,omitempty"`
	LocaliseKeyName string `json:"localise_key"`
}

// Response represents the search results
type Response struct {
	Count                 int           `json:"count"`
	Categories            []Category    `json:"categories"`
	Items                 []ContentItem `json:"items"`
	Suggestions           []string      `json:"suggestions,omitempty"`
	AdditionalSuggestions []string      `json:"additional_suggestions,omitempty"`
}

// Category represents all the search categories in search page
type Category struct {
	Count           int           `json:"count"`
	LocaliseKeyName string        `json:"localise_key"`
	ContentTypes    []ContentType `json:"content_types"`
}

// ContentType represents the type of the search results and the number of results for each type
type ContentType struct {
	Group           string   `json:"group"`
	Count           int      `json:"count"`
	LocaliseKeyName string   `json:"localise_key"`
	Types           []string `json:"types"`
}

// ContentItem represents each search result
type ContentItem struct {
	Type        ContentItemType `json:"type"`
	Description Description     `json:"description"`
	URI         string          `json:"uri"`
	Matches     *Matches        `json:"matches,omitempty"`
}

// ContentItemType represents the type of each search result
type ContentItemType struct {
	Type            string `json:"type"`
	LocaliseKeyName string `json:"localise_key"`
}

// Description represents each search result description
type Description struct {
	Contact           *Contact  `json:"contact,omitempty"`
	DatasetID         string    `json:"dataset_id,omitempty"`
	Edition           string    `json:"edition,omitempty"`
	Headline1         string    `json:"headline1,omitempty"`
	Headline2         string    `json:"headline2,omitempty"`
	Headline3         string    `json:"headline3,omitempty"`
	Keywords          *[]string `json:"keywords,omitempty"`
	LatestRelease     *bool     `json:"latest_release,omitempty"`
	Language          string    `json:"language,omitempty"`
	MetaDescription   string    `json:"meta_description,omitempty"`
	NationalStatistic *bool     `json:"national_statistic,omitempty"`
	NextRelease       string    `json:"next_release,omitempty"`
	PreUnit           string    `json:"pre_unit,omitempty"`
	ReleaseDate       string    `json:"release_date,omitempty"`
	Source            string    `json:"source,omitempty"`
	Summary           string    `json:"summary"`
	Title             string    `json:"title"`
	Unit              string    `json:"unit,omitempty"`
	Highlight         Highlight `json:"hightlight"`
}

// Hightlight contains specfic metadata with search keyword(s) highlighted
type Highlight struct {
	Title           string    `json:"title"`
	Keywords        *[]string `json:"keywords"`
	Summary         string    `json:"summary"`
	MetaDescription string    `json:"meta_description"`
	DatasetID       string    `json:"dataset_id"`
	Edition         string    `json:"edition"`
}

// Contact represents each search result contact details
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

// Matches represents each search result matches
type Matches struct {
	Description MatchDescription `json:"description"`
}

// MatchDescription represents each search result matches' description
type MatchDescription struct {
	Summary         *[]MatchDetails `json:"summary"`
	Title           *[]MatchDetails `json:"title"`
	Edition         *[]MatchDetails `json:"edition,omitempty"`
	MetaDescription *[]MatchDetails `json:"meta_description,omitempty"`
	Keywords        *[]MatchDetails `json:"keywords,omitempty"`
	DatasetID       *[]MatchDetails `json:"dataset_id,omitempty"`
}

// MatchDetails represents each search result matches' details
type MatchDetails struct {
	Value string `json:"value,omitempty"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Department represents other gov departmetns that match the search term
type Department struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Code  string `json:"code"`
	Match string `json:"match"`
}
