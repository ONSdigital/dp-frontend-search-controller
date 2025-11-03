package mapper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/model"

	helper "github.com/ONSdigital/dis-design-system-go/helper"
	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-frontend-search-controller/mocks"
	topicModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var mockTopicCategories = []data.Topic{
	{
		LocaliseKeyName:    "Census",
		Count:              1,
		DistinctItemsCount: 2,
		Query:              "1234",
		ShowInWebUI:        true,
	},
}

const (
	englishLang = "en"
	bindAddrAny = "localhost:0"
)

var expectedMappedBreadcrumb = []core.TaxonomyNode{
	{Title: "Home", URI: "/"},
	{Title: "Economy", URI: "/economy"},
	{Title: "Test", URI: "/economy/test"},
}

func TestUnitCreateSearchPage(t *testing.T) {
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.BindAddr = bindAddrAny
		req := httptest.NewRequest("", "/search", http.NoBody)
		mdl := core.Page{}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",

			Filter: data.Filter{
				Query:           []string{"article", "filter2", "publications"},
				LocaliseKeyName: []string{"Article"},
			},

			Sort: data.Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			CurrentPage: 1,
			TopicFilter: "1234",
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1
		categories[0].HideTypesInWebUI = true

		topicCategories := mockTopicCategories

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		Convey("When CreateSearchPage is called", func() {
			// NOTE: temporary measure until topic filter feature flag is removed
			cfg.EnableCensusTopicFilterOption = true

			sp := CreateSearchPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &topicModels.Navigation{}, []core.ErrorItem{})

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter, ShouldHaveLength, 3)
				So(sp.Data.Filter[0], ShouldEqual, "article")
				So(sp.Data.Filter[1], ShouldEqual, "filter2")
				So(sp.Data.Filter[2], ShouldEqual, "publications")

				So(sp.Data.Sort.Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.LocaliseFilterKeys, ShouldResemble, []string{"Article"})
				So(sp.Data.Sort.LocaliseSortKey, ShouldEqual, "Relevance")
				So(sp.Data.Sort.Options[0].Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.Options[0].LocaliseKeyName, ShouldEqual, "Relevance")

				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldEqual, "/search?q=housing&filter=article&filter=filter2&filter=publications&limit=10&sort=relevance&page=1")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.Categories, ShouldHaveLength, 3)
				So(sp.Data.Response.Categories[0].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 4)

				So(sp.Data.Response.Categories[0].ContentTypes[0].Group, ShouldEqual, "bulletin")
				So(sp.Data.Response.Categories[0].ContentTypes[0].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")

				So(sp.Data.Response.Categories[0].ContentTypes[1].Group, ShouldEqual, "article")
				So(sp.Data.Response.Categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].ContentTypes[1].LocaliseKeyName, ShouldEqual, "Article")

				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(sp.Data.Filters[0].FilterKey, ShouldHaveLength, 1)
				So(sp.Data.Filters[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Filters[0].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].NumberOfResults, ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[0].FilterKey, ShouldHaveLength, 1)
				So(sp.Data.Filters[0].Types[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")
				So(sp.Data.Filters[0].Types[0].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[0].NumberOfResults, ShouldEqual, 0)
				So(sp.Data.Filters[0].Types[1].FilterKey, ShouldHaveLength, 1)
				So(sp.Data.Filters[0].Types[1].LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Filters[0].Types[1].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].Types[1].NumberOfResults, ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[2].FilterKey, ShouldHaveLength, 1)
				So(sp.Data.Filters[0].Types[2].LocaliseKeyName, ShouldEqual, "Compendium")
				So(sp.Data.Filters[0].Types[2].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[2].NumberOfResults, ShouldEqual, 0)
				So(sp.Data.Filters[0].Types[3].FilterKey, ShouldHaveLength, 1)
				So(sp.Data.Filters[0].Types[3].LocaliseKeyName, ShouldEqual, "StatisticalArticle")
				So(sp.Data.Filters[0].Types[3].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[3].NumberOfResults, ShouldEqual, 0)
				So(sp.Data.Filters[2].Types, ShouldHaveLength, 3)

				So(sp.Data.TopicFilters, ShouldNotBeEmpty)
				So(sp.Data.TopicFilters[0], ShouldNotBeEmpty)
				So(sp.Data.TopicFilters[0].IsChecked, ShouldBeTrue)
				So(sp.Data.TopicFilters[0].LocaliseKeyName, ShouldEqual, "Census")
				So(sp.Data.TopicFilters[0].NumberOfResults, ShouldEqual, 1)
				So(sp.Data.TopicFilters[0].Query, ShouldEqual, "1234")
				So(sp.Data.TopicFilters[0].DistinctItemsCount, ShouldEqual, 2)

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)

				So(sp.SearchNoIndexEnabled, ShouldEqual, true)
				So(sp.Error.ErrorItems, ShouldHaveLength, 0)
			})
		})
	})
}

func TestUnitFindDatasetPage(t *testing.T) {
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.EnableCensusDimensionsFilterOption = true
		cfg.EnableCensusPopulationTypesFilterOption = true
		cfg.BindAddr = bindAddrAny
		req := httptest.NewRequest("GET", "/census/find-a-dataset", http.NoBody)
		mdl := core.Page{}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",

			Filter: data.Filter{
				Query:           []string{"dataset_landing_page"},
				LocaliseKeyName: []string{"Dataset"},
			},

			PopulationTypeFilter: "UR",

			DimensionsFilter: "ethnicity",

			Sort: data.Sort{
				Query:           "release_date",
				LocaliseKeyName: "Release Date",
			},

			Limit:       10,
			CurrentPage: 1,
			TopicFilter: "1234",
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		topicCategories := mockTopicCategories
		populationTypes := []data.PopulationTypes{}
		dimensions := []data.Dimensions{}

		respC, err := GetFindADatasetResponse()
		So(err, ShouldBeNil)

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		Convey("When CreateDataFinderPage is called", func() {
			// NOTE: temporary measure until topic filter feature flag is removed
			cfg.EnableCensusPopulationTypesFilterOption = true
			cfg.EnableCensusDimensionsFilterOption = true

			sp := CreateDataFinderPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, respC, englishLang, respH, []core.ErrorItem{}, &topicModels.Navigation{})

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter, ShouldHaveLength, 1)
				So(sp.Data.Filter[0], ShouldEqual, "dataset_landing_page")

				So(sp.Data.Sort.Query, ShouldEqual, "release_date")
				So(sp.Data.Sort.LocaliseFilterKeys, ShouldResemble, []string{"Dataset"})
				So(sp.Data.Sort.LocaliseSortKey, ShouldEqual, "Release Date")
				So(sp.Data.Sort.Options[0].Query, ShouldEqual, "release_date")
				So(sp.Data.Sort.Options[0].LocaliseKeyName, ShouldEqual, "ReleaseDate")

				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldEqual, "/census/find-a-dataset?dimensions=ethnicity&population_types=UR&q=housing&filter=dataset_landing_page&limit=10&sort=release_date&page=1")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.Categories, ShouldHaveLength, 3)
				So(sp.Data.Response.Categories[0].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 4)

				So(sp.Data.Response.Categories[0].ContentTypes[0].Group, ShouldEqual, "bulletin")
				So(sp.Data.Response.Categories[0].ContentTypes[0].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")

				So(sp.Data.Response.Categories[0].ContentTypes[1].Group, ShouldEqual, "article")
				So(sp.Data.Response.Categories[0].ContentTypes[1].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].ContentTypes[1].LocaliseKeyName, ShouldEqual, "Article")

				So(sp.Data.Response.Categories[0].ContentTypes[2].Group, ShouldEqual, "compendia")
				So(sp.Data.Response.Categories[0].ContentTypes[2].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[2].LocaliseKeyName, ShouldEqual, "Compendium")

				So(sp.Data.Response.Categories[0].ContentTypes[3].Group, ShouldEqual, "statistical_article")
				So(sp.Data.Response.Categories[0].ContentTypes[3].Count, ShouldEqual, 0)
				So(sp.Data.Response.Categories[0].ContentTypes[3].LocaliseKeyName, ShouldEqual, "StatisticalArticle")

				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "dataset_landing_page")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Datasets")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)
			})
		})
	})
}

func TestCreateDataAggregationPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.EnableAggregationPages = true
		cfg.BindAddr = bindAddrAny
		req := httptest.NewRequest("GET", "/alladhocs", http.NoBody)
		mdl := core.Page{}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",

			Sort: data.Sort{
				Query:           "release_date",
				LocaliseKeyName: "Release Date",
			},

			Limit:       10,
			CurrentPage: 1,
			TopicFilter: "1234",
		}

		categories := data.GetCategories()
		categories[0].Count = 1
		categories[0].ContentTypes[1].Count = 1

		topicCategories := mockTopicCategories

		respC, err := GetFindADatasetResponse()
		So(err, ShouldBeNil)

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		mockKeywordFilter := core.CompactSearch{
			ElementId: "keywords",
			InputName: "q",
			Language:  englishLang,
			Label: core.Localisation{
				LocaleKey: "SearchKeywords",
				Plural:    1,
			},
			SearchTerm: validatedQueryParams.Query,
		}

		lang := "en"
		mockAfterDate := core.DateFieldset{
			Language:                 lang,
			ValidationErrDescription: nil,
			ErrorID:                  validatedQueryParams.AfterDate.GetFieldsetErrID(),
			Input: core.InputDate{
				Language:              lang,
				Id:                    "after-date",
				InputNameDay:          "after-day",
				InputNameMonth:        "after-month",
				InputNameYear:         "after-year",
				InputValueDay:         validatedQueryParams.AfterDate.DayString(),
				InputValueMonth:       validatedQueryParams.AfterDate.MonthString(),
				InputValueYear:        validatedQueryParams.AfterDate.YearString(),
				HasDayValidationErr:   validatedQueryParams.AfterDate.HasDayValidationErr(),
				HasMonthValidationErr: validatedQueryParams.AfterDate.HasMonthValidationErr(),
				HasYearValidationErr:  validatedQueryParams.AfterDate.HasYearValidationErr(),
				DataAttributes: []core.DataAttribute{
					{
						Key: "invalid-date",
						Value: core.Localisation{
							LocaleKey: "ValidationInvalidDate",
							Plural:    1,
						},
					},
				},
				DayDataAttributes: []core.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "day"),
						},
					},
				},
				MonthDataAttributes: []core.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "month"),
						},
					},
				},
				YearDataAttributes: []core.DataAttribute{
					{
						Key: "value-missing",
						Value: core.Localisation{
							Text: helper.Localise("ValidationYearMissing", lang, 1, "after"),
						},
					},
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "year"),
						},
					},
				},
				Title: core.Localisation{
					LocaleKey: "ReleasedAfter",
					Plural:    1,
				},
				Description: core.Localisation{
					LocaleKey: "ReleasedAfterDescription",
					Plural:    1,
				},
			},
		}

		mockBeforeDate := core.DateFieldset{
			Language:                 lang,
			ValidationErrDescription: nil,
			ErrorID:                  validatedQueryParams.BeforeDate.GetFieldsetErrID(),
			Input: core.InputDate{
				Language:              lang,
				Id:                    "before-date",
				InputNameDay:          "before-day",
				InputNameMonth:        "before-month",
				InputNameYear:         "before-year",
				InputValueDay:         validatedQueryParams.BeforeDate.DayString(),
				InputValueMonth:       validatedQueryParams.BeforeDate.MonthString(),
				InputValueYear:        validatedQueryParams.BeforeDate.YearString(),
				HasDayValidationErr:   validatedQueryParams.BeforeDate.HasDayValidationErr(),
				HasMonthValidationErr: validatedQueryParams.BeforeDate.HasMonthValidationErr(),
				HasYearValidationErr:  validatedQueryParams.BeforeDate.HasYearValidationErr(),
				DataAttributes: []core.DataAttribute{
					{
						Key: "invalid-range",
						Value: core.Localisation{
							LocaleKey: "ValidationInvalidDateRange",
							Plural:    1,
						},
					},
					{
						Key: "invalid-date",
						Value: core.Localisation{
							LocaleKey: "ValidationInvalidDate",
							Plural:    1,
						},
					},
				},
				DayDataAttributes: []core.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "day"),
						},
					},
				},
				MonthDataAttributes: []core.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "month"),
						},
					},
				},
				YearDataAttributes: []core.DataAttribute{
					{
						Key: "value-missing",
						Value: core.Localisation{
							Text: helper.Localise("ValidationYearMissing", lang, 1, "before"),
						},
					},
					{
						Key: "pattern-mismatch",
						Value: core.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "year"),
						},
					},
				},
				Title: core.Localisation{
					LocaleKey: "ReleasedBefore",
					Plural:    1,
				},
				Description: core.Localisation{
					LocaleKey: "ReleasedBeforeDescription",
					Plural:    1,
				},
			},
		}

		Convey("When CreateDataAggregationPage is called", func() {
			sp := CreateDataAggregationPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, nil)

			Convey("Then successfully map core features to the page model", func() {
				// keyword search
				So(sp.Data.KeywordFilter, ShouldResemble, mockKeywordFilter)
				// date fieldsets
				So(sp.Data.BeforeDate, ShouldResemble, mockBeforeDate)
				So(sp.Data.AfterDate, ShouldResemble, mockAfterDate)
				// emergency banner
				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)
			})

			Convey("Then successfully map validation errors correctly to a page model", func() {
				validatedQueryParams.AfterDate = data.MustSetFieldsetErrID("fromDate-error")
				validatedQueryParams.BeforeDate = data.MustSetFieldsetErrID("toDate-error")

				validationErrs := []core.ErrorItem{
					{
						Description: core.Localisation{
							Text: "This is a released AFTER error",
						},
						ID:  "fromDate-error",
						URL: "#fromDate-error",
					},
					{
						Description: core.Localisation{
							Text: "This is a released BEFORE error",
						},
						ID:  "toDate-error",
						URL: "#toDate-error",
					},
					{
						Description: core.Localisation{
							Text: "This is another released BEFORE error",
						},
						ID:  "toDate-error",
						URL: "#toDate-error",
					},
					{
						Description: core.Localisation{
							Text: "This is a non-date page error",
						},
						ID:  "input-error",
						URL: "#input-error",
					},
				}

				expectedAfterErr := core.DateFieldset{
					ValidationErrDescription: []core.Localisation{
						{
							Text: validationErrs[0].Description.Text,
						},
					},
				}

				expectedBeforeErr := core.DateFieldset{
					ValidationErrDescription: []core.Localisation{
						{
							Text: validationErrs[1].Description.Text,
						},
						{
							Text: validationErrs[2].Description.Text,
						},
					},
				}

				page := CreateDataAggregationPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, validationErrs)
				So(page.Data.AfterDate.ValidationErrDescription, ShouldResemble, expectedAfterErr.ValidationErrDescription)
				So(page.Data.BeforeDate.ValidationErrDescription, ShouldResemble, expectedBeforeErr.ValidationErrDescription)
				So(page.Error.ErrorItems, ShouldResemble, validationErrs)
			})
		})

		Convey("When CreateDataAggregationPage is called with different page templates", func() {
			testcases := []struct {
				template                         string
				exTitle                          string
				exLocaliseKeyName                string
				exSingleContentTypeFilterEnabled bool
				exTopicFilterEnabled             bool
				exDateFilterEnabled              bool
				exEnableTimeSeriesExport         bool
				exRSSLink                        string
			}{
				{
					template:            "all-adhocs",
					exTitle:             "User requested data",
					exLocaliseKeyName:   "UserRequestedData",
					exDateFilterEnabled: true,
				},
				{
					template:                         "home-datalist",
					exTitle:                          "Published data",
					exLocaliseKeyName:                "DataList",
					exSingleContentTypeFilterEnabled: true,
					exDateFilterEnabled:              true,
					exRSSLink:                        "?rss",
				},
				{
					template:          "home-publications",
					exTitle:           "Publications",
					exLocaliseKeyName: "HomePublications",
					exRSSLink:         "?rss",
				},
				{
					template:             "all-methodologies",
					exTitle:              "All methodology",
					exLocaliseKeyName:    "AllMethodology",
					exTopicFilterEnabled: true,
				},
				{
					template:            "published-requests",
					exTitle:             "Freedom of Information (FOI) requests",
					exLocaliseKeyName:   "FOIRequests",
					exDateFilterEnabled: true,
				},
				{
					template:          "home-list",
					exTitle:           "Information pages",
					exLocaliseKeyName: "HomeList",
				},
				{
					template:          "home-methodology",
					exTitle:           "Methodology",
					exLocaliseKeyName: "HomeMethodology",
				},
				{
					template:                 "time-series-tool",
					exTitle:                  "Time series explorer",
					exLocaliseKeyName:        "TimeSeriesExplorer",
					exDateFilterEnabled:      true,
					exTopicFilterEnabled:     true,
					exEnableTimeSeriesExport: true,
				},
			}

			for _, tc := range testcases {
				Convey(fmt.Sprintf("Then page template: %s maps the page features correctly", tc.template), func() {
					sp := CreateDataAggregationPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &topicModels.Navigation{}, tc.template, cache.Topic{}, nil)
					So(sp.Metadata.Title, ShouldEqual, tc.exTitle)
					So(sp.Title.LocaliseKeyName, ShouldEqual, tc.exLocaliseKeyName)
					So(sp.Data.SingleContentTypeFilterEnabled, ShouldEqual, tc.exSingleContentTypeFilterEnabled)
					So(sp.Data.TopicFilterEnabled, ShouldEqual, tc.exTopicFilterEnabled)
					So(sp.Data.DateFilterEnabled, ShouldEqual, tc.exDateFilterEnabled)
					So(sp.RSSLink, ShouldEqual, tc.exRSSLink)
				})
			}
		})
	})
}

func TestCreatePreviousReleasesPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.BindAddr = bindAddrAny
		req := httptest.NewRequest("", "/foo/bar/previousreleases", http.NoBody)
		mdl := core.Page{}

		validatedQueryParams := data.SearchURLParams{
			Limit:       10,
			CurrentPage: 1,
		}

		respZ, err := GetMockZebedeePageDataResponse()
		So(err, ShouldBeNil)

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		respBc, err := GetMockBreadcrumbResponse()
		So(err, ShouldBeNil)

		Convey("When CreatePreviousReleasesPage is called", func() {
			sp := CreatePreviousReleasesPage(cfg, req, mdl, validatedQueryParams, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, nil, respZ, respBc)

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldStartWith, "/foo/bar/previousreleases")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)
				So(sp.Data.Response.Items, ShouldHaveLength, 1)

				So(sp.Data.Response.Items[0].Description.Keywords, ShouldHaveLength, 4)
				So(sp.Data.Response.Items[0].Description.MetaDescription, ShouldEqual, "Test Meta Description")
				So(sp.Data.Response.Items[0].Description.ReleaseDate, ShouldEqual, "2015-02-17T00:00:00.000Z")
				So(sp.Data.Response.Items[0].Description.Summary, ShouldEqual, "Test Summary")
				So(sp.Data.Response.Items[0].Description.Title, ShouldEqual, "Title Title")
				So(sp.Data.Response.Items[0].IsLatestRelease, ShouldBeTrue)

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Response.Items[0].URI, ShouldEqual, "/uri1/housing/articles/uri2/2015-02-17")

				So(sp.ServiceMessage, ShouldEqual, respH.ServiceMessage)

				So(len(sp.Breadcrumb), ShouldEqual, 4)
				So(sp.Breadcrumb[0], ShouldResemble, expectedMappedBreadcrumb[0])
				So(sp.Breadcrumb[1], ShouldResemble, expectedMappedBreadcrumb[1])
				So(sp.Breadcrumb[2], ShouldResemble, expectedMappedBreadcrumb[2])
				So(sp.Breadcrumb[3].Title, ShouldEqual, "Foo bar bulletin")
				So(sp.Breadcrumb[3].URI, ShouldEqual, "foo/bar/1/2/3")

				So(sp.EmergencyBanner.Type, ShouldEqual, strings.Replace(respH.EmergencyBanner.Type, "_", "-", -1))
				So(sp.EmergencyBanner.Title, ShouldEqual, respH.EmergencyBanner.Title)
				So(sp.EmergencyBanner.Description, ShouldEqual, respH.EmergencyBanner.Description)
				So(sp.EmergencyBanner.URI, ShouldEqual, respH.EmergencyBanner.URI)
				So(sp.EmergencyBanner.LinkText, ShouldEqual, respH.EmergencyBanner.LinkText)
			})
		})

		Convey("When CreatePreviousReleasesPage is called with validation errors", func() {
			validationErrs := []core.ErrorItem{
				{
					Description: core.Localisation{
						Text: "This is a current page error",
					},
					ID:  "currentPage-error",
					URL: "#currentPage-error",
				},
			}

			page := CreatePreviousReleasesPage(cfg, req, mdl, validatedQueryParams, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, validationErrs, respZ, respBc)

			Convey("Then validation errors are successfully mapped to the page model", func() {
				So(page.Error.ErrorItems, ShouldResemble, validationErrs)
			})
		})
	})
}

func TestMapLatestRelease(t *testing.T) {
	t.Parallel()

	Convey("When mapLatestRelease is called", t, func() {
		Convey("with a date that should match one item in response", func() {
			page := model.SearchPage{}
			searchResponse, _ := GetMockSearchResponse()
			mapResponse(&page, searchResponse, []data.Category{})
			latestReleaseDate := "2015-02-17T00:00:00.000Z"

			mapLatestRelease(&page, latestReleaseDate)

			Convey("then the matching item should be marked as the latest release", func() {
				So(page.Data.Response.Items[0].IsLatestRelease, ShouldBeTrue)
			})
		})
		Convey("with a date that doesn't match one item in response", func() {
			page := model.SearchPage{}
			searchResponse, _ := GetMockSearchResponse()
			mapResponse(&page, searchResponse, []data.Category{})
			latestReleaseDate := "2022-02-17T00:00:00.000Z"

			mapLatestRelease(&page, latestReleaseDate)

			Convey("then no item should be marked as the latest release", func() {
				So(page.Data.Response.Items[0].IsLatestRelease, ShouldBeFalse)
			})
		})
	})
}
