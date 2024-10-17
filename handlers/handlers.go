package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/dp-frontend-search-controller/model"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	searchAPI "github.com/ONSdigital/dp-search-api/api"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	searchError "github.com/ONSdigital/dp-search-api/sdk/errors"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

// Constants...
const (
	homepagePath = "/"
	Bearer       = "Bearer "
)

// list of content types that have /previousreleases
var knownPreviousReleaseTypes = []string{
	"bulletin",
	"article",
	"compendium_landing_page",
}

// list of query params allowed on /previousreleases
var allowedPreviousReleasesQueryParams = []string{data.Page}

// HandlerClients represents the handlers for search and data-aggregation
type HandlerClients struct {
	Renderer      RenderClient
	SearchClient  SearchClient
	ZebedeeClient ZebedeeClient
	TopicClient   TopicClient
}

// NewHandlerClients creates a new instance of FilterFlex
func NewHandlerClients(rc RenderClient, sc SearchClient, zc ZebedeeClient, tc TopicClient) *HandlerClients {
	return &HandlerClients{
		Renderer:      rc,
		SearchClient:  sc,
		ZebedeeClient: zc,
		TopicClient:   tc,
	}
}

// Read Handler
func Read(cfg *config.Config, hc *HandlerClients, cacheList cache.List, template string) http.HandlerFunc {
	oldHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList, template, false)
	})

	newHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList, template, true)
	})

	return cookies.Handler(cfg.ABTest.Enabled, newHandler, oldHandler, cfg.ABTest.Percentage, cfg.ABTest.AspectID, cfg.SiteDomain, cfg.ABTest.Exit)
}

// ReadDataAggregationWithTopics for data aggregation routes with topic/subtopics
func ReadDataAggregationWithTopics(cfg *config.Config, hc *HandlerClients, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregationWithTopics(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList, template)
	})
}

func ReadDataAggregation(cfg *config.Config, hc *HandlerClients, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregation(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList, template)
	})
}

// ReadPreviousReleases handles previous releases page
func ReadPreviousReleases(cfg *config.Config, hc *HandlerClients, cacheList cache.List) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readPreviousReleases(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList)
	})
}

// ReadRelated data handles related data page
func ReadRelatedData(cfg *config.Config, hc *HandlerClients, cacheList cache.List) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readRelatedData(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList)
	})
}

func ReadFindDataset(cfg *config.Config, hc *HandlerClients, cacheList cache.List) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readFindDataset(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList)
	})
}

func readFindDataset(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	urlQuery := req.URL.Query()
	urlQuery.Del("filter")
	urlQuery.Add("filter", "dataset_landing_page")
	urlQuery.Add("filter", "user_requested_data")

	// get cached census topic and its subtopics
	censusTopicCache, err := cacheList.CensusTopic.GetCensusData(ctx)
	if err != nil {
		log.Error(ctx, "failed to get census topic cache for dataset", err)
		setStatusCode(w, req, err)
		return
	}

	clearTopics := false
	if urlQuery.Get("topics") == "" {
		urlQuery.Add("topics", censusTopicCache.Query)
		clearTopics = true
	}

	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache for dataset", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, err := data.ReviewDatasetQuery(ctx, cfg, urlQuery, censusTopicCache)
	if err != nil && !apperrors.ErrMapForRenderBeforeAPICalls[err] {
		log.Error(ctx, "unable to review dataset query", err)
		setStatusCode(w, req, err)
		return
	}

	var (
		// counter used to keep track of the number of concurrent API calls
		counter            = 3
		errorMessage       string
		makeSearchAPICalls = true
	)

	// avoid making unnecessary search API calls
	if apperrors.ErrMapForRenderBeforeAPICalls[err] {
		makeSearchAPICalls = false

		// reduce counter by the number of concurrent search API calls that would be
		// run in go routines
		counter -= 2
		errorMessage = err.Error()
	}

	searchQuery := data.GetSearchAPIQuery(validatedQueryParams, censusTopicCache)
	categoriesCountQuery := getCategoriesDatasetCountQuery(searchQuery)

	var (
		homepageResp zebedeeCli.HomepageContent
		searchResp   = &searchModels.SearchResponse{}

		categories      []data.Category
		topicCategories []data.Topic
		populationTypes []data.PopulationTypes
		dimensions      []data.Dimensions

		wg sync.WaitGroup

		respErr, countErr error
	)

	wg.Add(counter)

	go func() {
		defer wg.Done()
		var homeErr error
		homepageResp, homeErr = zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
		if homeErr != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}))
			return
		}
	}()
	if makeSearchAPICalls {
		var options searchSDK.Options

		options.Query = searchQuery
		options.Headers = http.Header{
			searchSDK.CollectionID: {collectionID},
		}

		setFlorenceTokenHeader(options.Headers, accessToken)

		go func() {
			defer wg.Done()

			searchResp, respErr = searchC.GetSearch(ctx, options)
			if respErr != nil {
				log.Error(ctx, "getting search response from client failed for dataset", respErr)
				cancel()
				return
			}
		}()
		go func() {
			defer wg.Done()

			// TO-DO: Need to make a second request until API can handle aggregation on datatypes (e.g. bulletins, article) to return counts
			categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
			if countErr != nil {
				log.Error(ctx, "getting categories, types and its counts failed for dataset", countErr)
				setStatusCode(w, req, countErr)
				cancel()
				return
			}
		}()
	}

	wg.Wait()
	if respErr != nil || countErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	if clearTopics {
		/* By default, we set all topics as active,
		 * but we don't want the checkboxes to be ticked
		 * this ensures they're sent to the topic API, but
		 * hides that from the frontend.
		 */
		validatedQueryParams.TopicFilter = ""
	}

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		log.Error(ctx, "unable to validate current page for dataset", err)
		setStatusCode(w, req, err)
		return
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateDataFinderPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, searchResp, lang, homepageResp, errorMessage, navigationCache)
	rend.BuildPage(w, m, "search")
}

func readDataAggregation(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List, template string,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	var err error
	urlQuery := req.URL.Query()

	// get cached census topic and its subtopics
	censusTopicCache, err := cacheList.CensusTopic.GetCensusData(ctx)
	if err != nil {
		log.Error(ctx, "failed to get census topic cache for aggregation", err)
		setStatusCode(w, req, err)
		return
	}

	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache for aggregation", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, validationErrs := data.ReviewDataAggregationQueryWithParams(ctx, cfg, urlQuery)
	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed for aggregation", log.Data{
			"parameters": validationErrs,
		})
		// Errors are now mapped to the page model to output feedback to the user rather than
		// a blank 400 error response.
		m := mapper.CreateDataAggregationPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs)
		buildDataAggregationPage(w, m, rend, template)
		return
	}

	if _, rssParam := urlQuery["rss"]; rssParam {
		req.Header.Set("Accept", "application/rss+xml")
		if err = createRSSFeed(ctx, w, req, collectionID, accessToken, searchC, validatedQueryParams, template); err != nil {
			log.Error(ctx, "failed to create rss feed for aggregation", err)
			setStatusCode(w, req, err)
			return
		}
		return
	}

	// counter used to keep track of the number of concurrent API calls
	var counter = 3

	searchQuery := data.GetDataAggregationQuery(validatedQueryParams, template)
	categoriesCountQuery := getCategoriesCountQuery(searchQuery)

	var (
		homepageResp zebedeeCli.HomepageContent
		searchResp   = &searchModels.SearchResponse{}

		categories      []data.Category
		topicCategories []data.Topic

		wg sync.WaitGroup

		respErr, countErr error
	)
	wg.Add(counter)

	go func() {
		defer wg.Done()
		var homeErr error
		homepageResp, homeErr = zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
		if homeErr != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}))
			return
		}
	}()

	var options searchSDK.Options

	options.Query = searchQuery

	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setFlorenceTokenHeader(options.Headers, accessToken)

	go func() {
		defer wg.Done()

		searchResp, respErr = searchC.GetSearch(ctx, options)
		if respErr != nil {
			log.Error(ctx, "getting search response from client failed for aggregation", respErr)
			cancel()
			return
		}
	}()

	go func() {
		defer wg.Done()

		// TO-DO: Need to make a second request until API can handle aggregation on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
		if countErr != nil {
			log.Error(ctx, "getting categories, types and its counts failed for aggregation", countErr)
			setStatusCode(w, req, countErr)
			cancel()
			return
		}
	}()

	wg.Wait()
	if respErr != nil || countErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: "current page exceeds total pages",
			},
		})
		m := mapper.CreateDataAggregationPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs)
		buildDataAggregationPage(w, m, rend, template)
		return
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateDataAggregationPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs)
	buildDataAggregationPage(w, m, rend, template)
}

func readPreviousReleases(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	var err error
	template := "related-list-pages"
	urlPath := path.Dir(req.URL.Path)
	urlQuery := req.URL.Query()
	latestContentURL := urlPath + "/latest"

	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)
	// check page type
	pageData, err := zc.GetPageData(ctx, accessToken, collectionID, lang, latestContentURL)
	if err != nil {
		log.Error(ctx, "failed to get content type", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !slices.Contains(knownPreviousReleaseTypes, pageData.Type) {
		err := errors.New("page type doesn't match known list of content types compatible with /previousreleases")
		log.Error(ctx, "page type isn't compatible with /previousreleases", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache for aggregation", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, validationErrs := data.ReviewPreviousReleasesQueryWithParams(ctx, cfg, sanitisedParams, urlPath)

	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed for aggregation", log.Data{
			"parameters": validationErrs,
		})
		// Errors are now mapped to the page model to output feedback to the user rather than
		// a blank 400 error response.
		m := mapper.CreatePreviousReleasesPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs, zebedeeCli.PageData{}, []zebedeeCli.Breadcrumb{})
		buildDataAggregationPage(w, m, rend, template)
		return
	}

	// counter used to keep track of the number of concurrent API calls
	var counter = 3
	searchQuery := data.SetParentTypeOnSearchAPIQuery(validatedQueryParams, pageData.Type)

	var (
		homepageResp zebedeeCli.HomepageContent
		searchResp   = &searchModels.SearchResponse{}
		bc           []zebedeeCli.Breadcrumb

		wg sync.WaitGroup

		respErr, countErr, bcErr error
	)
	wg.Add(counter)

	go func() {
		defer wg.Done()
		var homeErr error
		homepageResp, homeErr = zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
		if homeErr != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{homeErr}))
			return
		}
	}()

	var options searchSDK.Options

	options.Query = searchQuery

	options.Headers = http.Header{
		searchSDK.FlorenceToken: {Bearer + accessToken},
		searchSDK.CollectionID:  {collectionID},
	}

	go func() {
		defer wg.Done()

		searchResp, respErr = searchC.GetSearch(ctx, options)
		if respErr != nil {
			log.Error(ctx, "getting search response from client failed for aggregation", respErr)
			cancel()
			return
		}
	}()

	go func() {
		defer wg.Done()

		bc, bcErr = zc.GetBreadcrumb(ctx, accessToken, collectionID, lang, latestContentURL)
		if bcErr != nil {
			bc = []zebedeeCli.Breadcrumb{}
			return
		}
	}()

	wg.Wait()
	if respErr != nil || countErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	basePage := rend.NewBasePageModel()
	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: "current page exceeds total pages",
			},
		})
		m := mapper.CreatePreviousReleasesPage(cfg, req, basePage, validatedQueryParams, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
		rend.BuildPage(w, m, template)
		return
	}

	m := mapper.CreatePreviousReleasesPage(cfg, req, basePage, validatedQueryParams, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
	rend.BuildPage(w, m, template)
}

// Maps template name to underlying go template
func buildDataAggregationPage(w http.ResponseWriter, m model.SearchPage, rend RenderClient, template string) {
	// time-series-tool needs its own template due to the need of elements to be present for JS to be able to assign onClick events(doesn't work if they're conditionally shown on the page)
	if template != "time-series-tool" {
		rend.BuildPage(w, m, "data-aggregation-page")
	} else {
		rend.BuildPage(w, m, template)
	}
}

func readDataAggregationWithTopics(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List, template string,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	// Capture the path after the prefix
	vars := mux.Vars(req)
	topicsPath := vars["topicsPath"]

	// Split the remaining path into segments
	segments := strings.Split(topicsPath, "/")

	selectedTopic := cache.Topic{}

	// Validate the topic hierarchy
	lastSegmentTopic, err := ValidateTopicHierarchy(ctx, segments, cacheList)
	if err != nil {
		log.Error(ctx, "invalid topic path", err, log.Data{
			"topicPath": topicsPath,
		})
		err = apperrors.ErrTopicPathNotFound
		setStatusCode(w, req, err)
		return
	}

	selectedTopic = *lastSegmentTopic

	urlQuery := req.URL.Query()
	urlQuery.Add("topics", selectedTopic.ID)

	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache for aggregation with topics", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, validationErrs := data.ReviewDataAggregationQueryWithParams(ctx, cfg, urlQuery)
	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed", log.Data{
			"parameters": validationErrs,
		})
		// Errors are now mapped to the page model to output feedback to the user rather than
		// a blank 400 error response.
		m := mapper.CreateDataAggregationPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs)
		buildDataAggregationPage(w, m, rend, template)
		return
	}

	if _, rssParam := urlQuery["rss"]; rssParam {
		req.Header.Set("Accept", "application/rss+xml")
		if err = createRSSFeed(ctx, w, req, collectionID, accessToken, searchC, validatedQueryParams, template); err != nil {
			log.Error(ctx, "failed to create rss feed with topics", err)
			setStatusCode(w, req, err)
			return
		}
		return
	}
	// counter used to keep track of the number of concurrent API calls
	var counter = 3

	searchQuery := data.GetDataAggregationQuery(validatedQueryParams, template)
	categoriesCountQuery := getCategoriesTopicsCountQuery(searchQuery)

	var (
		homepageResp zebedeeCli.HomepageContent
		searchResp   = &searchModels.SearchResponse{}

		categories      []data.Category
		topicCategories []data.Topic

		wg sync.WaitGroup

		respErr, countErr error
	)
	wg.Add(counter)

	go func() {
		defer wg.Done()
		var homeErr error
		homepageResp, homeErr = zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
		if homeErr != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}))
			return
		}
	}()

	var options searchSDK.Options

	options.Query = searchQuery

	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setFlorenceTokenHeader(options.Headers, accessToken)

	go func() {
		defer wg.Done()

		searchResp, respErr = searchC.GetSearch(ctx, options)
		if respErr != nil {
			log.Error(ctx, "getting search response from client failed for aggregation with topics", respErr)
			cancel()
			return
		}
	}()

	go func() {
		defer wg.Done()

		// TO-DO: Need to make a second request until API can handle aggregation on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, &selectedTopic)
		if countErr != nil {
			log.Error(ctx, "getting categories, types and its counts failed for aggregation with topics", countErr)
			setStatusCode(w, req, countErr)
			cancel()
			return
		}
	}()

	wg.Wait()
	if respErr != nil || countErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		log.Error(ctx, "unable to validate current page for aggregation with topics", err)
		setStatusCode(w, req, err)
		return
	}
	basePage := rend.NewBasePageModel()
	m := mapper.CreateDataAggregationPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, searchResp, lang, homepageResp, "", navigationCache, template, selectedTopic, validationErrs)
	// time-series-tool needs its own template due to the need of elements to be present for JS to be able to assign onClick events(doesn't work if they're conditionally shown on the page)
	if template != "time-series-tool" {
		rend.BuildPage(w, m, "data-aggregation-page")
	} else {
		rend.BuildPage(w, m, template)
	}
}

func read(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List, template string, nlpWeightingEnabled bool,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	urlQuery := req.URL.Query()

	// get cached census topic and its subtopics
	censusTopicCache, err := cacheList.CensusTopic.GetCensusData(ctx)
	if err != nil {
		log.Error(ctx, "failed to get census topic cache", err)
		setStatusCode(w, req, err)
		return
	}

	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, validationErrs := data.ReviewQuery(ctx, cfg, urlQuery, censusTopicCache)
	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed", log.Data{
			"parameters": validationErrs,
		})
		basePage := rend.NewBasePageModel()
		m := mapper.CreateSearchPage(cfg, req, basePage, validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, validationErrs)
		rend.BuildPage(w, m, template)
		return
	}

	validatedQueryParams.NLPWeightingEnabled = nlpWeightingEnabled
	log.Info(ctx, "NLP Weighting for query", log.Data{
		"nlp_weighting": nlpWeightingEnabled,
	})

	var (
		// counter used to keep track of the number of concurrent API calls
		counter            = 3
		errorMessage       string
		makeSearchAPICalls = true
	)

	searchQuery := data.GetSearchAPIQuery(validatedQueryParams, censusTopicCache)
	categoriesCountQuery := getCategoriesCountQuery(searchQuery)

	var (
		homepageResp zebedeeCli.HomepageContent
		searchResp   = &searchModels.SearchResponse{}

		categories      []data.Category
		topicCategories []data.Topic
		wg              sync.WaitGroup

		respErr, countErr error
	)
	wg.Add(counter)

	go func() {
		defer wg.Done()
		var homeErr error
		homepageResp, homeErr = zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
		if homeErr != nil {
			log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}))
			return
		}
	}()

	if makeSearchAPICalls {
		var options searchSDK.Options

		options.Query = searchQuery

		options.Headers = http.Header{
			searchSDK.CollectionID: {collectionID},
		}

		setFlorenceTokenHeader(options.Headers, accessToken)

		go func() {
			defer wg.Done()

			searchResp, respErr = searchC.GetSearch(ctx, options)
			if respErr != nil {
				log.Error(ctx, "getting search response from client failed", respErr)
				cancel()
				return
			}
		}()
		go func() {
			defer wg.Done()

			// TO-DO: Need to make a second request until API can handle aggregation on datatypes (e.g. bulletins, article) to return counts
			categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
			if countErr != nil {
				log.Error(ctx, "getting categories, types and its counts failed", countErr)
				setStatusCode(w, req, countErr)
				cancel()
				return
			}
		}()
	}

	wg.Wait()
	if respErr != nil || countErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: apperrors.ErrPageExceedsTotalPages.Error(),
			},
		})
	}
	basePage := rend.NewBasePageModel()
	m := mapper.CreateSearchPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, searchResp, lang, homepageResp, errorMessage, navigationCache, validationErrs)
	rend.BuildPage(w, m, template)
}

// validateCurrentPage checks if the current page exceeds the total pages which is a bad request
func validateCurrentPage(ctx context.Context, cfg *config.Config, validatedQueryParams data.SearchURLParams, resultsCount int) error {
	if resultsCount > 0 {
		totalPages := data.GetTotalPages(cfg, validatedQueryParams.Limit, resultsCount)

		if validatedQueryParams.CurrentPage > totalPages {
			err := apperrors.ErrPageExceedsTotalPages
			log.Error(ctx, "current page exceeds total pages", err)

			return err
		}
	}

	return nil
}

// getCategoriesCountQuery removes specific params to return the total count for all types.
func getCategoriesCountQuery(searchQuery url.Values) url.Values {
	return removeQueryParams(searchQuery, "content_type", "topics", "population_types", "dimensions")
}

// getCategoriesTopicsCountQuery removes fewer params, for counts based on topics.
func getCategoriesTopicsCountQuery(searchQuery url.Values) url.Values {
	return removeQueryParams(searchQuery, "content_type", "population_types", "dimensions")
}

// getCategoriesDatasetCountQuery removes a different set of params for dataset counts.
func getCategoriesDatasetCountQuery(searchQuery url.Values) url.Values {
	return removeQueryParams(searchQuery, "topics", "population_types", "dimensions")
}

// removeQueryParams clones the search query and removes specified params.
func removeQueryParams(searchQuery url.Values, paramsToRemove ...string) url.Values {
	// Clone the searchQuery to avoid modifying the original copy
	query := url.Values(http.Header(searchQuery).Clone())
	// Remove specified params
	for _, param := range paramsToRemove {
		query.Del(param)
	}

	return query
}

// getCategoriesTypesCount removes the filters and communicates with the search api again to retrieve the number of search results for each filter categories and subtypes
func getCategoriesTypesCount(ctx context.Context, accessToken, collectionID string, query url.Values, searchC SearchClient, topicCache *cache.Topic) ([]data.Category, []data.Topic, error) {
	var options searchSDK.Options

	options.Query = query
	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setFlorenceTokenHeader(options.Headers, accessToken)

	countResp, err := searchC.GetSearch(ctx, options)
	if err != nil {
		logData := log.Data{"url_values": query}
		log.Error(ctx, "getting search query count from client failed", err, logData)
		return nil, nil, err
	}

	categories := data.GetCategories()
	topicCategories := data.GetTopics(topicCache, countResp)

	setCountToCategories(ctx, countResp, categories)

	return categories, topicCategories, nil
}

func setCountToCategories(ctx context.Context, countResp *searchModels.SearchResponse, categories []data.Category) {
	for _, responseType := range countResp.ContentTypes {
		foundFilter := false

	categoryLoop:
		for i, category := range categories {
			for j, contentType := range category.ContentTypes {
				for _, subType := range contentType.Types {
					if responseType.Type == subType {
						categories[i].Count += responseType.Count
						categories[i].ContentTypes[j].Count += responseType.Count

						foundFilter = true

						break categoryLoop
					}
				}
			}
		}

		if !foundFilter && !data.IsCategoryUnused(responseType.Type) {
			log.Warn(ctx, "unrecognised filter type returned from api", log.Data{"filter_type": responseType.Type})
		}
	}
}

func setStatusCode(w http.ResponseWriter, req *http.Request, err error) {
	status := http.StatusInternalServerError

	// Zebedee returns a ClientError interface
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}

	// Search API returns a SearchClientError interface
	if err, ok := err.(SearchClientError); ok {
		if err.Status() == http.StatusBadRequest {
			status = err.Status()
		}
	}

	if apperrors.BadRequestMap[err] {
		status = http.StatusBadRequest
	}

	if apperrors.NotFoundMap[err] {
		status = http.StatusNotFound
	}

	log.Error(req.Context(), "setting-response-status", err)

	w.WriteHeader(status)
}

func setFlorenceTokenHeader(headers http.Header, accessToken string) {
	if strings.HasPrefix(accessToken, Bearer) {
		headers.Set(searchSDK.FlorenceToken, accessToken)
	} else {
		headers.Set(searchSDK.FlorenceToken, Bearer+accessToken)
	}
}

func createRSSFeed(ctx context.Context, w http.ResponseWriter, r *http.Request, collectionID, accessToken string, api SearchClient, validatedParams data.SearchURLParams, template string) error {
	var err error
	uriPrefix := "https://www.ons.gov.uk"

	var options searchSDK.Options

	options.Query = data.GetDataAggregationQuery(validatedParams, template)

	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setFlorenceTokenHeader(options.Headers, accessToken)

	searchResponse, respErr := api.GetSearch(ctx, options)
	if respErr != nil {
		log.Error(ctx, "getting search response from client for rss feed failed", respErr)
		setStatusCode(w, r, respErr)
		return respErr
	}

	w.Header().Set("Content-Type", "application/rss+xml")

	pageTitle, pageTag := getPageTitle(template)

	feed := &feeds.Feed{
		Title: pageTitle + " - Office for National Statistics",
		Link:  &feeds.Link{Href: "https://www.ons.gov.uk/" + pageTag},
	}

	feed.Items = []*feeds.Item{}
	for i := range searchResponse.Items {
		resp := &searchResponse.Items[i]
		date, parseErr := time.Parse("2006-01-02T15:04:05.000Z", resp.ReleaseDate)
		if parseErr != nil {
			return fmt.Errorf("error parsing time: %s", parseErr)
		}
		item := &feeds.Item{
			Title:       resp.Title,
			Link:        &feeds.Link{Href: uriPrefix + resp.URI},
			Description: resp.Summary,
			Id:          uriPrefix + resp.URI,
			Created:     date,
		}

		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Error(ctx, "Error converting feed to RSS", err)
		return fmt.Errorf("error converting to rss: %s", err)
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write([]byte(rss))
	if err != nil {
		log.Error(ctx, "Error writing RSS to response", err)
		return fmt.Errorf("error writing rss to response: %s", err)
	}

	return nil
}

func getPageTitle(template string) (pageTitle, pageTag string) {
	switch template {
	case "all-adhocs":
		return "User requested data", "UserRequestedData"
	case "home-datalist":
		return "Published data", "DataList"
	case "home-publications":
		return "Publications", "HomePublications"
	case "all-methodologies":
		return "All methodology", "AllMethodology"
	case "published-requests":
		return "Freedom of Information (FOI) requests", "FOIRequests"
	case "home-list":
		return "Information pages", "HomeList"
	case "home-methodology":
		return "Methodology", "HomeMethodology"
	case "time-series-tool":
		return "Time series explorer", "TimeSeriesExplorer"
	}

	return "", ""
}

// ValidateTopicHierarchy validate the segments i.e. check that they all exist in the cache, check that the hierarchy is correct and return the last item as the selectedTopic
func ValidateTopicHierarchy(ctx context.Context, segments []string, cacheList cache.List) (*cache.Topic, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments to validate")
	}

	// Start with the first segment
	currentTopic, err := cacheList.DataTopic.GetTopic(ctx, segments[0], "")
	if err != nil {
		return nil, fmt.Errorf("invalid topic hierarchy at segment: %s", segments[0])
	}

	// Traverse through segments
	for i := 1; i < len(segments); i++ {
		nextTopic, err := cacheList.DataTopic.GetTopic(ctx, segments[i], currentTopic.Slug)
		if err != nil || nextTopic.ParentID != currentTopic.ID {
			return nil, fmt.Errorf("invalid topic hierarchy at segment: %s", segments[i])
		}
		currentTopic = nextTopic
	}

	return cacheList.DataTopic.GetTopicFromSubtopic(currentTopic), nil
}

// sanitiseQueryParams takes a predefined list of allowed query params and removes any from the request URL that don't match
func sanitiseQueryParams(allowedParams []string, inputParams url.Values) url.Values {
	sanitisedParams := url.Values{}
	for paramKey, paramValue := range inputParams {
		for _, allowedParam := range allowedParams {
			if paramKey == allowedParam {
				for _, param := range paramValue {
					sanitisedParams.Add(paramKey, param)
				}
			}
		}
	}
	return sanitisedParams
}

// checkAllowedPageTypes calls Zebedee for a given URL and checks if it's page type matches against a list of allowed page types
func checkAllowedPageTypes(ctx context.Context, w http.ResponseWriter, zc ZebedeeClient, accessToken, collectionID, lang, pageURL string, allowedPagewTypes []string) (zebedeeCli.PageData, error) {
	pageData, err := zc.GetPageData(ctx, accessToken, collectionID, lang, pageURL)
	if err != nil {
		log.Error(ctx, "failed to get content type", err)
		return zebedeeCli.PageData{}, err
	}
	if !slices.Contains(knownPreviousReleaseTypes, pageData.Type) {
		err := errors.New("page type doesn't match known list of content types compatible with /previousreleases")
		log.Error(ctx, "page type isn't compatible with /previousreleases", err)
		return zebedeeCli.PageData{}, err
	}
	return pageData, nil
}

// getNavigationCache returns cached navigation data
func getNavigationCache(ctx context.Context, w http.ResponseWriter, req *http.Request, cacheList cache.List, lang string) *models.Navigation {
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache for aggregation", err)
		setStatusCode(w, req, err)
	}
	return navigationCache
}

// getSearch performs a get request to search API
func getSearch(ctx context.Context, searchC SearchClient, options searchSDK.Options, cancel func()) (*searchModels.SearchResponse, searchError.Error) {
	s, err := searchC.GetSearch(ctx, options)
	if err != nil {
		log.Error(ctx, "getting search response from client failed", err)
		cancel()
		return nil, err
	}
	return s, nil
}

// postSearchURIs posts a list of URIs to search API and gets a search response
func postSearchURIs(ctx context.Context, searchC SearchClient, options searchSDK.Options, cancel func(), URIsRequest searchAPI.URIsRequest) (*searchModels.SearchResponse, searchError.Error) {
	if len(URIsRequest.URIs) > 0 {
		s, err := searchC.PostSearchURIs(ctx, options, URIsRequest)
		if err != nil {
			log.Error(ctx, "getting search response from client failed", err)
			cancel()
			return nil, err
		}
		return s, nil
	}
	return nil, nil
}

// getBreadcrumb performs a get request to zebedee for breadcrumb data
func getBreadcrumb(ctx context.Context, zc ZebedeeClient, accessToken, collectionID, lang, pageURL string) []zebedeeCli.Breadcrumb {
	bc, err := zc.GetBreadcrumb(ctx, accessToken, collectionID, lang, pageURL)
	if err != nil {
		log.Warn(ctx, "getting breadcrumb response from client failed", log.FormatErrors([]error{err}))
		bc = []zebedeeCli.Breadcrumb{}
	}
	return bc
}

// getHomepageContent performs a get request to zebedee for breadcrumb data
func getHomepageContent(ctx context.Context, zc ZebedeeClient, accessToken, collectionID, lang string) zebedeeCli.HomepageContent {
	hp, err := zc.GetHomepageContent(ctx, accessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "getting homepage response from client failed", log.FormatErrors([]error{err}))
		hp = zebedeeCli.HomepageContent{}
	}
	return hp
}
