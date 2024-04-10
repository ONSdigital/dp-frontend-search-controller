package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	topicModels "github.com/ONSdigital/dp-topic-api/models"
	topicSDK "github.com/ONSdigital/dp-topic-api/sdk"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/feeds"
)

// Constants...
const (
	homepagePath = "/"
)

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

// Read Handler for data aggregation routes with topic/subtopics
func ReadDataAggregationWithTopics(cfg *config.Config, hc *HandlerClients, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregationWithTopics(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, hc.TopicClient, accessToken, collectionID, lang, cacheList, template)
	})
}

// Read Handler
func ReadDataAggregation(cfg *config.Config, hc *HandlerClients, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregation(w, req, cfg, hc.ZebedeeClient, hc.Renderer, hc.SearchClient, accessToken, collectionID, lang, cacheList, template)
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
		log.Error(ctx, "failed to get census topic cache", err)
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
		log.Error(ctx, "failed to get navigation cache", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, err := data.ReviewDatasetQuery(ctx, cfg, urlQuery, censusTopicCache)
	if err != nil && !apperrors.ErrMapForRenderBeforeAPICalls[err] {
		log.Error(ctx, "unable to review query", err)
		setStatusCode(w, req, err)
		return
	}

	var (
		// counter used to keep track of the number of concurrent API calls
		counter            = 3
		errorMessage       string
		makeSearchAPICalls = true
	)

	// avoid making unecessary search API calls
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
				log.Error(ctx, "getting search response from client failed", respErr)
				cancel()
				return
			}
		}()
		go func() {
			defer wg.Done()

			// TO-DO: Need to make a second request until API can handle aggregration on datatypes (e.g. bulletins, article) to return counts
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

	if clearTopics {
		/* By default we set all topics as active,
		 * but we don't want the checkboxes to be ticked
		 * this ensures they're sent to the topic API, but
		 * hides that from the frontend.
		 */
		validatedQueryParams.TopicFilter = ""
	}

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		log.Error(ctx, "unable to validate current page", err)
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

	validatedQueryParams, err := data.ReviewDataAggregationQuery(ctx, cfg, urlQuery, censusTopicCache)
	if err != nil && !apperrors.ErrMapForRenderBeforeAPICalls[err] {
		log.Error(ctx, "unable to review query", err)
		setStatusCode(w, req, err)
		return
	}

	if _, rssParam := urlQuery["rss"]; rssParam {
		req.Header.Set("Accept", "application/rss+xml")
		if err = createRSSFeed(ctx, w, req, collectionID, accessToken, searchC, validatedQueryParams, template); err != nil {
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

	var options searchSDK.Options

	options.Query = searchQuery

	options.Headers = http.Header{
		searchSDK.FlorenceToken: {"Bearer " + accessToken},
		searchSDK.CollectionID:  {collectionID},
	}

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

		// TO-DO: Need to make a second request until API can handle aggregration on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
		if countErr != nil {
			log.Error(ctx, "getting categories, types and its counts failed", countErr)
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
		log.Error(ctx, "unable to validate current page", err)
		setStatusCode(w, req, err)
		return
	}
	basePage := rend.NewBasePageModel()
	m := mapper.CreateDataAggregationPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, searchResp, lang, homepageResp, "", navigationCache, template, topicModels.Topic{})
	// time-series-tool needs it's own template due to the need of elements to be present for JS to be able to assign onClick events(doesn't work if they're conditionally shown on the page)
	if template != "time-series-tool" {
		rend.BuildPage(w, m, "data-aggregation-page")
	} else {
		rend.BuildPage(w, m, template)
	}
}

func readDataAggregationWithTopics(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient, topicC TopicClient,
	accessToken, collectionID, lang string, cacheList cache.List, template string,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	vars := mux.Vars(req)

	respRootTopics, topicAPIError := topicC.GetRootTopicsPublic(ctx, topicSDK.Headers{})
	if topicAPIError != nil {
		logData := log.Data{
			"req_headers": topicSDK.Headers{},
		}
		log.Error(ctx, "failed to get root topics from topic api", topicAPIError, logData)
		return
	}

	rootTopicItems := *respRootTopics.PublicItems
	selectedTopic := topicModels.Topic{}

	topicPath := vars["topic"]
	topic, err := getTopicByURLString(topicPath, rootTopicItems)
	if err != nil {
		log.Error(ctx, "could not match topicPath to topics", err, log.Data{
			"topicPath": topicPath,
		})
		setStatusCode(w, req, err)
		return
	}

	subtopicPath := vars["subTopic"]
	if subtopicPath != "" {
		subTopics, topicAPIError := topicC.GetSubtopicsPublic(ctx, topicSDK.Headers{}, topic.ID)
		if topicAPIError != nil {
			log.Error(ctx, "failed to get subtopics", topicAPIError)
			setStatusCode(w, req, topicAPIError)
			return
		}

		subtopicItems := *subTopics.PublicItems

		subtopic, matchingErr := getTopicByURLString(subtopicPath, subtopicItems)
		if matchingErr != nil {
			log.Error(ctx, "could not match subtopicPath to subtopics", matchingErr, log.Data{
				"subtopicPath": subtopicPath,
			})
			setStatusCode(w, req, matchingErr)
			return
		}

		selectedTopic = subtopic
	} else {
		selectedTopic = topic
	}

	urlQuery := req.URL.Query()

	urlQuery.Add("topics", selectedTopic.ID)

	// replace with new cache
	censusTopicCache, err := cacheList.CensusTopic.GetCensusData(ctx)
	if err != nil {
		log.Error(ctx, "failed to get census topic cache", err)
		setStatusCode(w, req, err)
		return
	}

	log.Info(ctx, "this the census cache topics", log.Data{"topics": censusTopicCache})

	// get cached navigation data
	navigationCache, err := cacheList.Navigation.GetNavigationData(ctx, lang)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache", err)
		setStatusCode(w, req, err)
		return
	}

	validatedQueryParams, err := data.ReviewDataAggregationQueryWithParams(ctx, cfg, urlQuery, censusTopicCache)
	if err != nil && !apperrors.ErrMapForRenderBeforeAPICalls[err] {
		log.Error(ctx, "unable to review query", err)
		setStatusCode(w, req, err)
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

	var options searchSDK.Options

	options.Query = searchQuery

	options.Headers = http.Header{
		searchSDK.FlorenceToken: {"Bearer " + accessToken},
		searchSDK.CollectionID:  {collectionID},
	}

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

		// TO-DO: Need to make a second request until API can handle aggregration on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
		if countErr != nil {
			log.Error(ctx, "getting categories, types and its counts failed", countErr)
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
		log.Error(ctx, "unable to validate current page", err)
		setStatusCode(w, req, err)
		return
	}
	basePage := rend.NewBasePageModel()
	m := mapper.CreateDataAggregationPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, searchResp, lang, homepageResp, "", navigationCache, template, selectedTopic)
	// time-series-tool needs it's own template due to the need of elements to be present for JS to be able to assign onClick events(doesn't work if they're conditionally shown on the page)
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

	validatedQueryParams, err := data.ReviewQuery(ctx, cfg, urlQuery, censusTopicCache)
	if err != nil && !apperrors.ErrMapForRenderBeforeAPICalls[err] {
		log.Error(ctx, "unable to review query", err)
		setStatusCode(w, req, err)
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

	// avoid making unecessary search API calls
	if apperrors.ErrMapForRenderBeforeAPICalls[err] {
		makeSearchAPICalls = false

		// reduce counter by the number of concurrent search API calls that would be
		// run in go routines
		counter -= 2
		errorMessage = err.Error()
	}

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

			// TO-DO: Need to make a second request until API can handle aggregration on datatypes (e.g. bulletins, article) to return counts
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
		log.Error(ctx, "unable to validate current page", err)
		setStatusCode(w, req, err)
		return
	}
	basePage := rend.NewBasePageModel()
	m := mapper.CreateSearchPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, searchResp, lang, homepageResp, errorMessage, navigationCache)
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

// getCategoriesCountQuery clones url (query) values before removing
// filters to be able to return total counts for different filters
func getCategoriesCountQuery(searchQuery url.Values) url.Values {
	// Clone the searchQuery url values to prevent changing the original copy
	query := url.Values(http.Header(searchQuery).Clone())

	// Remove filter to get count of all types for the query from the client
	query.Del("content_type")
	query.Del("topics")
	query.Del("population_types")
	query.Del("dimensions")

	return query
}

// getCategoriesCountQuery clones url (query) values before removing
// filters to be able to return total counts for different filters
func getCategoriesDatasetCountQuery(searchQuery url.Values) url.Values {
	// Clone the searchQuery url values to prevent changing the original copy
	query := url.Values(http.Header(searchQuery).Clone())

	// Remove filter to get count of all types for the query from the client
	query.Del("topics")
	query.Del("population_types")
	query.Del("dimensions")

	return query
}

// getCategoriesTypesCount removes the filters and communicates with the search api again to retrieve the number of search results for each filter categories and subtypes
func getCategoriesTypesCount(ctx context.Context, accessToken, collectionID string, query url.Values, searchC SearchClient, censusTopicCache *cache.Topic) ([]data.Category, []data.Topic, error) {
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
	topicCategories := data.GetTopics(censusTopicCache, countResp)

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

		if !foundFilter {
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
	if strings.HasPrefix(accessToken, "Bearer ") {
		headers.Set(searchSDK.FlorenceToken, accessToken)
	} else {
		headers.Set(searchSDK.FlorenceToken, "Bearer "+accessToken)
	}
}

func createRSSFeed(ctx context.Context, w http.ResponseWriter, r *http.Request, collectionID, accessToken string, api SearchClient, validatedParams data.SearchURLParams, template string) error {
	var err error
	uriPrefix := "https://www.ons.gov.uk"

	var options searchSDK.Options

	options.Query = data.GetDataAggregationQuery(validatedParams, template)

	options.Headers = http.Header{
		searchSDK.FlorenceToken: {"Bearer " + accessToken},
		searchSDK.CollectionID:  {collectionID},
	}

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

// getTopicByURLString matches a URL string, e.g. businessindustryandtrade against
// a Topic retrieved from the Topic API, using it's Title attribute, e.g.
// "Business, industry and trade"
func getTopicByURLString(topicURLString string, topics []topicModels.Topic) (topicModels.Topic, error) {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	for _, topic := range topics {
		if nonAlphanumericRegex.ReplaceAllString(strings.ToLower(topic.Title), "") == strings.ToLower(topicURLString) {
			return topic, nil
		}
	}
	return topicModels.Topic{}, apperrors.ErrTopicPathNotFound
}
