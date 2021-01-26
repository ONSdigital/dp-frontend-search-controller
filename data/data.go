package data

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
