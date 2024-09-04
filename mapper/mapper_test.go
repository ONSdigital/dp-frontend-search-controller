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

	"github.com/ONSdigital/dp-frontend-search-controller/mocks"
	helper "github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
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

const englishLang string = "en"

func TestUnitCreateSearchPage(t *testing.T) {
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		req := httptest.NewRequest("", "/search", http.NoBody)
		mdl := coreModel.Page{}

		validatedQueryParams := data.SearchURLParams{
			Query: "housing",

			Filter: data.Filter{
				Query:           []string{"article", "filter2"},
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

		topicCategories := mockTopicCategories

		respH, err := GetMockHomepageContent()
		So(err, ShouldBeNil)

		respC, err := GetMockSearchResponse()
		So(err, ShouldBeNil)

		Convey("When CreateSearchPage is called", func() {
			// NOTE: temporary measure until topic filter feature flag is removed
			cfg.EnableCensusTopicFilterOption = true

			sp := CreateSearchPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, respC, englishLang, respH, "", &topicModels.Navigation{})

			Convey("Then successfully map search response from search-query client to page model", func() {
				So(sp.Data.Query, ShouldEqual, "housing")
				So(sp.Data.Filter, ShouldHaveLength, 2)
				So(sp.Data.Filter[0], ShouldEqual, "article")
				So(sp.Data.Filter[1], ShouldEqual, "filter2")

				So(sp.Data.Sort.Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.LocaliseFilterKeys, ShouldResemble, []string{"Article"})
				So(sp.Data.Sort.LocaliseSortKey, ShouldEqual, "Relevance")
				So(sp.Data.Sort.Options[0].Query, ShouldEqual, "relevance")
				So(sp.Data.Sort.Options[0].LocaliseKeyName, ShouldEqual, "Relevance")

				So(sp.Data.Pagination.CurrentPage, ShouldEqual, 1)
				So(sp.Data.Pagination.TotalPages, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay, ShouldHaveLength, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].PageNumber, ShouldEqual, 1)
				So(sp.Data.Pagination.PagesToDisplay[0].URL, ShouldEqual, "/search?q=housing&filter=article&filter=filter2&limit=10&sort=relevance&page=1")
				So(sp.Data.Pagination.Limit, ShouldEqual, 10)
				So(sp.Data.Pagination.LimitOptions, ShouldResemble, []int{10, 25, 50})

				So(sp.Data.Response.Count, ShouldEqual, 1)

				So(sp.Data.Response.Categories, ShouldHaveLength, 3)
				So(sp.Data.Response.Categories[0].Count, ShouldEqual, 1)
				So(sp.Data.Response.Categories[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 3)

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

				So(len(sp.Data.Filters[0].FilterKey), ShouldEqual, 3)
				So(sp.Data.Filters[0].LocaliseKeyName, ShouldEqual, "Publication")
				So(sp.Data.Filters[0].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.Filters[0].Types[0].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[0].LocaliseKeyName, ShouldEqual, "StatisticalBulletin")
				So(sp.Data.Filters[0].Types[0].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[0].NumberOfResults, ShouldEqual, 0)
				So(len(sp.Data.Filters[0].Types[1].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[1].LocaliseKeyName, ShouldEqual, "Article")
				So(sp.Data.Filters[0].Types[1].IsChecked, ShouldBeTrue)
				So(sp.Data.Filters[0].Types[1].NumberOfResults, ShouldEqual, 1)
				So(len(sp.Data.Filters[0].Types[2].FilterKey), ShouldEqual, 1)
				So(sp.Data.Filters[0].Types[2].LocaliseKeyName, ShouldEqual, "Compendium")
				So(sp.Data.Filters[0].Types[2].IsChecked, ShouldBeFalse)
				So(sp.Data.Filters[0].Types[2].NumberOfResults, ShouldEqual, 0)
				So(len(sp.Data.Filters[2].Types), ShouldEqual, 3)

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
			})
		})
	})
}

func TestUnitFindDatasetPage(t *testing.T) {
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		cfg.EnableCensusDimensionsFilterOption = true
		cfg.EnableCensusPopulationTypesFilterOption = true
		So(err, ShouldBeNil)
		req := httptest.NewRequest("GET", "/census/find-a-dataset", http.NoBody)
		mdl := coreModel.Page{}

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

			sp := CreateDataFinderPage(cfg, req, mdl, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, respC, englishLang, respH, "", &topicModels.Navigation{})

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
				So(sp.Data.Response.Categories[0].ContentTypes, ShouldHaveLength, 3)

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
	t.Parallel()
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		cfg.EnableAggregationPages = true
		So(err, ShouldBeNil)
		req := httptest.NewRequest("GET", "/alladhocs", http.NoBody)
		mdl := coreModel.Page{}

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

		mockKeywordFilter := coreModel.CompactSearch{
			ElementId: "keywords",
			InputName: "q",
			Language:  englishLang,
			Label: coreModel.Localisation{
				LocaleKey: "SearchKeywords",
				Plural:    1,
			},
			SearchTerm: validatedQueryParams.Query,
		}

		lang := "en"
		mockAfterDate := coreModel.DateFieldset{
			Language:                 lang,
			ValidationErrDescription: nil,
			ErrorID:                  validatedQueryParams.AfterDate.GetFieldsetErrID(),
			Input: coreModel.InputDate{
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
				DataAttributes: []coreModel.DataAttribute{
					{
						Key: "invalid-date",
						Value: coreModel.Localisation{
							LocaleKey: "ValidationInvalidDate",
							Plural:    1,
						},
					},
				},
				DayDataAttributes: []coreModel.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "day"),
						},
					},
				},
				MonthDataAttributes: []coreModel.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "month"),
						},
					},
				},
				YearDataAttributes: []coreModel.DataAttribute{
					{
						Key: "value-missing",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationYearMissing", lang, 1, "after"),
						},
					},
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "after", "year"),
						},
					},
				},
				Title: coreModel.Localisation{
					LocaleKey: "ReleasedAfter",
					Plural:    1,
				},
				Description: coreModel.Localisation{
					LocaleKey: "ReleasedAfterDescription",
					Plural:    1,
				},
			},
		}

		mockBeforeDate := coreModel.DateFieldset{
			Language:                 lang,
			ValidationErrDescription: nil,
			ErrorID:                  validatedQueryParams.BeforeDate.GetFieldsetErrID(),
			Input: coreModel.InputDate{
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
				DataAttributes: []coreModel.DataAttribute{
					{
						Key: "invalid-range",
						Value: coreModel.Localisation{
							LocaleKey: "ValidationInvalidDateRange",
							Plural:    1,
						},
					},
					{
						Key: "invalid-date",
						Value: coreModel.Localisation{
							LocaleKey: "ValidationInvalidDate",
							Plural:    1,
						},
					},
				},
				DayDataAttributes: []coreModel.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "day"),
						},
					},
				},
				MonthDataAttributes: []coreModel.DataAttribute{
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "month"),
						},
					},
				},
				YearDataAttributes: []coreModel.DataAttribute{
					{
						Key: "value-missing",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationYearMissing", lang, 1, "before"),
						},
					},
					{
						Key: "pattern-mismatch",
						Value: coreModel.Localisation{
							Text: helper.Localise("ValidationPatternMismatch", lang, 1, "before", "year"),
						},
					},
				},
				Title: coreModel.Localisation{
					LocaleKey: "ReleasedBefore",
					Plural:    1,
				},
				Description: coreModel.Localisation{
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

				validationErrs := []coreModel.ErrorItem{
					{
						Description: coreModel.Localisation{
							Text: "This is a released AFTER error",
						},
						ID:  "fromDate-error",
						URL: "#fromDate-error",
					},
					{
						Description: coreModel.Localisation{
							Text: "This is a released BEFORE error",
						},
						ID:  "toDate-error",
						URL: "#toDate-error",
					},
					{
						Description: coreModel.Localisation{
							Text: "This is another released BEFORE error",
						},
						ID:  "toDate-error",
						URL: "#toDate-error",
					},
					{
						Description: coreModel.Localisation{
							Text: "This is a non-date page error",
						},
						ID:  "input-error",
						URL: "#input-error",
					},
				}

				expectedAfterErr := coreModel.DateFieldset{
					ValidationErrDescription: []coreModel.Localisation{
						{
							Text: validationErrs[0].Description.Text,
						},
					},
				}

				expectedBeforeErr := coreModel.DateFieldset{
					ValidationErrDescription: []coreModel.Localisation{
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
					template:                         "home-publications",
					exTitle:                          "Publications",
					exLocaliseKeyName:                "HomePublications",
					exSingleContentTypeFilterEnabled: true,
					exRSSLink:                        "?rss",
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
	t.Parallel()

	Convey("Given validated query and response from search-api", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		req := httptest.NewRequest("", "/foo/bar/previousreleases", http.NoBody)
		mdl := coreModel.Page{}

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

		Convey("When CreatePreviousReleasesPage is called", func() {
			sp := CreatePreviousReleasesPage(cfg, req, mdl, validatedQueryParams, respC, englishLang, respH, "", &topicModels.Navigation{}, "", cache.Topic{}, nil, respZ)

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

				So(sp.Data.Response.Items[0].Type.Type, ShouldEqual, "article")
				So(sp.Data.Response.Items[0].Type.LocaliseKeyName, ShouldEqual, "Article")
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
