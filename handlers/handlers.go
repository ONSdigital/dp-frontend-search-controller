package handlers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"
)

// Constants...
const (
	homepagePath = "/"
)

// Read Handler
func Read(cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, zc, rend, searchC, accessToken, collectionID, lang, cacheList, template)
	})
}

// Read Handler
func ReadDataAggregation(cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient, cacheList cache.List, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregation(w, req, cfg, zc, rend, searchC, accessToken, collectionID, lang, cacheList, template)
	})
}

func ReadFindDataset(cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient, cacheList cache.List) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readFindDataset(w, req, cfg, zc, rend, searchC, accessToken, collectionID, lang, cacheList)
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
	if err != nil && !errs.ErrMapForRenderBeforeAPICalls[err] {
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
	if errs.ErrMapForRenderBeforeAPICalls[err] {
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

	if err != nil && !errs.ErrMapForRenderBeforeAPICalls[err] {
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
		log.Info(ctx, "searchResp", log.Data{"searchResp": searchResp})
		if respErr != nil {
			log.Error(ctx, "getting search response from client failed", respErr)
			cancel()
			return
		}
	}()

	go func() {
		defer wg.Done()

		// TO-DO: Need to make a second request until API can handle aggregration on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, populationTypes, dimensions, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, censusTopicCache)
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
	m := mapper.CreateDataAggregationPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, searchResp, lang, homepageResp, "", navigationCache, template)
	rend.BuildPage(w, m, template)
}

func read(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List, template string,
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

	if err != nil && !errs.ErrMapForRenderBeforeAPICalls[err] {
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
	if errs.ErrMapForRenderBeforeAPICalls[err] {
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
	makeSearchAPICalls = true
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
	m := mapper.CreateSearchPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, populationTypes, dimensions, searchResp, lang, homepageResp, errorMessage, navigationCache)
	rend.BuildPage(w, m, template)
}

// validateCurrentPage checks if the current page exceeds the total pages which is a bad request
func validateCurrentPage(ctx context.Context, cfg *config.Config, validatedQueryParams data.SearchURLParams, resultsCount int) error {
	if resultsCount > 0 {
		totalPages := data.GetTotalPages(cfg, validatedQueryParams.Limit, resultsCount)

		if validatedQueryParams.CurrentPage > totalPages {
			err := errs.ErrPageExceedsTotalPages
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

	if errs.BadRequestMap[err] {
		status = http.StatusBadRequest
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
