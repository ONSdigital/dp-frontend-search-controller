package model

import (
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Search is the model struct for the cookies preferences form
type SearchPage struct {
	model.Page
	Data       Search          `json:"data"`
	Title      Title           `json:"title,omitempty"`
	BeforeDate model.InputDate `json:"before_date"`
	AfterDate  model.InputDate `json:"after_date"`
	RSSLink    string          `json:"rss_link"`
}

// Search represents all search parameters and response data of the search
type Search struct {
	Query                          string                 `json:"query"`
	ErrorMessage                   string                 `json:"error_message,omitempty"`
	EnabledFilters                 []string               `json:"enabled_filters,omitempty"`
	DateFilterEnabled              bool                   `json:"data_filter_enabled,omitempty"`
	EnableTimeSeriesExport         bool                   `json:"enable_time_series_export,omitempty"`
	TopicFilterEnabled             bool                   `json:"topic_filter_enabled,omitempty"`
	KeywordFilter                  model.CompactSearch    `json:"keyword_filter"`
	ContentTypeFilterEnabled       bool                   `json:"content_type_filter_enabled,omitempty"`
	SingleContentTypeFilterEnabled bool                   `json:"single_content_type_filter_enabled,omitempty"`
	Filter                         []string               `json:"filter,omitempty"`
	Filters                        []Filter               `json:"filters"`
	BeforeDate                     model.DateFieldset     `json:"before_date"`
	AfterDate                      model.DateFieldset     `json:"after_date"`
	TopicFilters                   []TopicFilter          `json:"topic_filters"`
	CensusFilters                  []TopicFilter          `json:"census_filters"`
	PopulationTypeFilter           []PopulationTypeFilter `json:"population_types"`
	DimensionsFilter               []DimensionsFilter     `json:"dimensions"`
	Sort                           Sort                   `json:"sort,omitempty"`
	Pagination                     model.Pagination       `json:"pagination,omitempty"`
	Response                       Response               `json:"response"`
	TermLocalKey                   string                 `json:"term_localise_key_name,omitempty"`
	Topic                          string                 `json:"topic,omitempty"`
	FeedbackAPIURL                 string                 `json:"feedback_api_url"`
}

// Filter represents all filter information needed by templates
type Filter struct {
	LocaliseKeyName string   `json:"localise_key_name,omitempty"`
	FilterKey       []string `json:"filter_key,omitempty"`
	IsChecked       bool     `json:"is_checked,omitempty"`
	NumberOfResults int      `json:"number_of_results,omitempty"`
	Types           []Filter `json:"types,omitempty"`
	HideTypes       bool     `json:"hide_types,omitempty"`
}

// TopicFilter represents all the topic filter information needed by templates
type TopicFilter struct {
	LocaliseKeyName    string        `json:"localise_key_name,omitempty"`
	DistinctItemsCount int           `json:"distinct_items_count,omitempty"`
	Query              string        `json:"query,omitempty"`
	IsChecked          bool          `json:"is_checked,omitempty"`
	NumberOfResults    int           `json:"number_of_results,omitempty"`
	Types              []TopicFilter `json:"subtopics,omitempty"`
}

type PopulationTypeFilter struct {
	LocaliseKeyName string `json:"localise_key_name,omitempty"`
	Count           int    `json:"count,omitempty"`
	Query           string `json:"query,omitempty"`
	IsChecked       bool   `json:"is_checked,omitempty"`
	NumberOfResults int    `json:"number_of_results,omitempty"`
	Type            string `json:"type,omitempty"`
}

type DimensionsFilter struct {
	LocaliseKeyName string `json:"localise_key_name,omitempty"`
	Count           int    `json:"count,omitempty"`
	Query           string `json:"query,omitempty"`
	IsChecked       bool   `json:"is_checked,omitempty"`
	NumberOfResults int    `json:"number_of_results,omitempty"`
	Type            string `json:"type,omitempty"`
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
	Type            ContentItemType `json:"type"`
	Dataset         Dataset         `json:"dataset"`
	Description     Description     `json:"description"`
	URI             string          `json:"uri"`
	Matches         *Matches        `json:"matches,omitempty"`
	IsLatestRelease bool            `json:"is_latest_release"`
}

// ContentItemType represents the type of each search result
type ContentItemType struct {
	Type            string `json:"type"`
	LocaliseKeyName string `json:"localise_key"`
}

// Dataset represents additional dataset fields
type Dataset struct {
	PopulationType string `json:"population_type,omitempty"`
}

// Description represents each search result description
type Description struct {
	Contact           *Contact  `json:"contact,omitempty"`
	CDID              string    `json:"cdid,omitempty"`
	DatasetID         string    `json:"dataset_id,omitempty"`
	Edition           string    `json:"edition,omitempty"`
	Headline1         string    `json:"headline1,omitempty"`
	Headline2         string    `json:"headline2,omitempty"`
	Headline3         string    `json:"headline3,omitempty"`
	Keywords          []string  `json:"keywords,omitempty"`
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
	Keywords        []*string `json:"keywords"`
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

type Title struct {
	Title           string `json:"title"`
	LocaliseKeyName string `json:"localise_key"`
}
