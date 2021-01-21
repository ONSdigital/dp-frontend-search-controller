package data

// FilterType informs the name of the search type displayed on the website, the query retrieved from renderer and all the subtypes to pass to the logic
type FilterType struct {
	Name      string   `bson:"name" json:"name"`
	QueryType string   `bson:"query_type" json:"query_type"`
	SubTypes  []string `bson:"sub_types" json:"sub_types"`
}

// Category contains the high level search categories and it's corresponding search types
// If search types are added or removed in the map, make sure to do the same in the defaultContentTypes variable in dp-search-query
var Category = map[string][]FilterType{
	"Publication": {Bulletin, Article, Compendium},
	"Data":        {TimeSeries, Datasets, UserRequestedData},
	"Other":       {Methodology, CorporateInformation},
}

// Bulletin - Search information specific for statistical bulletins
var Bulletin = FilterType{
	Name:      "Statistical bulletin",
	QueryType: "bulletin",
	SubTypes:  []string{"bulletin"},
}

// Article - Search information specific for articles
var Article = FilterType{
	Name:      "Article",
	QueryType: "article",
	SubTypes:  []string{"article", "article_download"},
}

// Compendium - Search information specific for compendium
var Compendium = FilterType{
	Name:      "Compendium",
	QueryType: "compendia",
	SubTypes:  []string{"compendium_landing_page"},
}

// TimeSeries - Search information specific for time series
var TimeSeries = FilterType{
	Name:      "Time series",
	QueryType: "time_series",
	SubTypes:  []string{"timeseries"},
}

// Datasets - Search information specific for datasets
var Datasets = FilterType{
	Name:      "Datasets",
	QueryType: "datasets",
	SubTypes:  []string{"dataset_landing_page", "reference_tables"},
}

// UserRequestedData - Search information specific for user requested data
var UserRequestedData = FilterType{
	Name:      "User requested data",
	QueryType: "user_requested_data",
	SubTypes:  []string{"static_adhoc"},
}

// Methodology - Search information specific for methodologies
var Methodology = FilterType{
	Name:      "Methodology",
	QueryType: "methodology",
	SubTypes:  []string{"static_methodology", "static_methodology_download", "static_qmi"},
}

// CorporateInformation - Search information specific for corporate information
var CorporateInformation = FilterType{
	Name:      "Corporate Information",
	QueryType: "corporate_information",
	SubTypes:  []string{"static_foi", "static_page", "static_landing_page", "static_article"},
}
