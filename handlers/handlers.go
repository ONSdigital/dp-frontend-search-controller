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
	"github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/model"
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
	homepagePath         = "/"
	Bearer               = "Bearer "
	RelatedPagesTemplate = "related-list-pages"
)

// list of content types that have /relateddata and /previousreleases
var knownRelatedListTypes = []string{
	"bulletin",
	"article",
	"compendium_landing_page",
}

// list of query params allowed on /previousreleases
var allowedPreviousReleasesQueryParams = []string{data.Page}

// SearchHandler represents the handlers for search functionality
type SearchHandler struct {
	Renderer                    RenderClient
	SearchClient                SearchClient
	TopicClient                 TopicClient
	ZebedeeClient               ZebedeeClient
	EnableAggregationPages      bool
	EnableTopicAggregationPages bool
	CacheList                   cache.List
}

// NewSearchHandler creates a new instance of SearchHandler
func NewSearchHandler(rc RenderClient, sc SearchClient, tc TopicClient, zc ZebedeeClient, cfg *config.Config, cl cache.List) *SearchHandler {
	return &SearchHandler{
		Renderer:                    rc,
		SearchClient:                sc,
		TopicClient:                 tc,
		ZebedeeClient:               zc,
		EnableAggregationPages:      cfg.EnableAggregationPages,
		EnableTopicAggregationPages: cfg.EnableTopicAggregationPages,
		CacheList:                   cl,
	}
}

type AggregationConfig struct {
	TemplateName                       string
	URLQueryParams                     url.Values
	UseTopicsPath                      bool
	UseURIsRequest                     bool
	NLPWeightingEnabled                bool
	ValidateParams                     func(context.Context, *config.Config, url.Values, string, *cache.Topic) (data.SearchURLParams, []core.ErrorItem)
	CreatePageModel                    func(*config.Config, *http.Request, core.Page, data.SearchURLParams, []data.Category, []data.Topic, *searchModels.SearchResponse, string, zebedeeCli.HomepageContent, string, *models.Navigation, string, cache.Topic, []core.ErrorItem, zebedeeCli.PageData, []zebedeeCli.Breadcrumb) model.SearchPage
	GetSearchAndCategoriesCountQueries func(data.SearchURLParams, *cache.Topic, string, string) (url.Values, url.Values)
}

func handleReadRequest(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient, accessToken, collectionID, lang string, cacheList cache.List, aggCfg AggregationConfig) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	urlPath := path.Dir(req.URL.Path)
	latestContentURL := path.Dir(req.URL.Path) + "/latest"
	// Extract topic if required
	var selectedTopic cache.Topic

	selectedTopic, err := selectTopic(ctx, req, cacheList, aggCfg)
	if err != nil {
		setStatusCode(w, req, err)
		return
	}

	clearTopics := prepareQueryParams(req, &aggCfg, selectedTopic)
	pageData, err := validatePageType(ctx, w, zc, aggCfg, accessToken, collectionID, lang, urlPath)
	if err != nil {
		setStatusCode(w, req, err)
		return
	}
	counter := 2
	wg := sync.WaitGroup{}

	var homepageResp zebedeeCli.HomepageContent
	var navigationCache *models.Navigation
	var searchResp = &searchModels.SearchResponse{}
	var categories []data.Category
	var topicCategories []data.Topic
	var respErr, countErr error
	var searchCount int
	var bc []zebedeeCli.Breadcrumb

	wg.Add(counter)

	// Parallel fetching
	go func() {
		defer wg.Done()
		// get cached navigation data
		navigationCache = getNavigationCache(ctx, w, req, cacheList, lang)
	}()
	go func() {
		defer wg.Done()
		// get homepage content
		homepageResp = getHomepageContent(ctx, zc, accessToken, collectionID, lang)
	}()

	if aggCfg.TemplateName == RelatedPagesTemplate {
		wg.Add(1)
		if aggCfg.UseURIsRequest {
			go func() {
				defer wg.Done()
				bc = getBreadcrumb(ctx, zc, accessToken, collectionID, lang, urlPath)
			}()
		} else {
			go func() {
				defer wg.Done()
				bc = getBreadcrumb(ctx, zc, accessToken, collectionID, lang, latestContentURL)
			}()
		}
	}
	wg.Wait()

	validatedQueryParams, validationErrs := aggCfg.ValidateParams(ctx, cfg, aggCfg.URLQueryParams, urlPath, &selectedTopic)
	if len(validationErrs) > 0 {
		m := aggCfg.CreatePageModel(cfg, req, rend.NewBasePageModel(), validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, aggCfg.TemplateName, selectedTopic, validationErrs, pageData, []zebedeeCli.Breadcrumb{})
		buildDataAggregationPage(w, m, rend, aggCfg.TemplateName)
		return
	}

	if aggCfg.NLPWeightingEnabled {
		validatedQueryParams.NLPWeightingEnabled = aggCfg.NLPWeightingEnabled
		log.Info(ctx, "NLP Weighting for query", log.Data{
			"nlp_weighting": aggCfg.NLPWeightingEnabled,
		})
	}

	if _, rssParam := aggCfg.URLQueryParams["rss"]; rssParam {
		var err error
		req.Header.Set("Accept", "application/rss+xml")
		if err = createRSSFeed(ctx, w, req, collectionID, accessToken, searchC, validatedQueryParams, aggCfg.TemplateName); err != nil {
			log.Error(ctx, "failed to create rss feed", err)
			setStatusCode(w, req, err)
		}
		return
	}

	searchQuery, categoriesCountQuery := aggCfg.GetSearchAndCategoriesCountQueries(validatedQueryParams, &selectedTopic, aggCfg.TemplateName, pageData.Type)

	var options searchSDK.Options

	options.Query = searchQuery
	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setAuthTokenHeader(options.Headers, accessToken)
	if aggCfg.TemplateName == RelatedPagesTemplate {
		if aggCfg.UseURIsRequest {
			URIList := make([]string, 0, len(pageData.RelatedData))
			for _, related := range pageData.RelatedData {
				URIList = append(URIList, related.URI)
			}

			URIsRequest := searchAPI.URIsRequest{
				URIs:   URIList,
				Limit:  validatedQueryParams.Limit,
				Offset: validatedQueryParams.Offset,
				Sort:   validatedQueryParams.Sort.Query,
			}

			searchResp, respErr, searchCount = postSearchURIs(ctx, searchC, options, cancel, URIsRequest)
			if respErr != nil {
				log.Error(ctx, "getting search response with uris from client failed", respErr)
			}
		} else {
			searchResp, respErr = searchC.GetSearch(ctx, options)
			if respErr != nil {
				log.Error(ctx, "getting search response from client failed for dataset", respErr)
			}
			searchCount = searchResp.Count
		}
	} else {
		searchResp, respErr = searchC.GetSearch(ctx, options)
		if respErr != nil {
			log.Error(ctx, "getting search response from client failed for dataset", respErr)
		}
		searchCount = searchResp.Count

		// TO-DO: Need to make a second request until API can handle aggregation on datatypes (e.g. bulletins, article) to return counts
		categories, topicCategories, countErr = getCategoriesTypesCount(ctx, accessToken, collectionID, categoriesCountQuery, searchC, &selectedTopic)
		if countErr != nil {
			log.Error(ctx, "getting categories, types and its counts failed for dataset", countErr)
			setStatusCode(w, req, countErr)
		}
	}
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

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchCount)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: err.Error(),
			},
		})
		m := aggCfg.CreatePageModel(cfg, req, rend.NewBasePageModel(), validatedQueryParams, []data.Category{}, []data.Topic{}, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, aggCfg.TemplateName, cache.Topic{}, validationErrs, pageData, bc)
		buildDataAggregationPage(w, m, rend, aggCfg.TemplateName)
		return
	}

	m := aggCfg.CreatePageModel(cfg, req, rend.NewBasePageModel(), validatedQueryParams, categories, topicCategories, searchResp, lang, homepageResp, "", navigationCache, aggCfg.TemplateName, selectedTopic, validationErrs, pageData, bc)
	buildDataAggregationPage(w, m, rend, aggCfg.TemplateName)
}

func getSelectedTopic(ctx context.Context, req *http.Request, cacheList cache.List) (cache.Topic, error) {
	vars := mux.Vars(req)
	topicsPath := vars["topicsPath"]

	// Split the remaining path into segments
	segments := strings.Split(topicsPath, "/")

	lastSegmentTopic, err := ValidateTopicHierarchy(ctx, segments, cacheList)
	if err != nil {
		log.Error(ctx, "invalid topic path", err, log.Data{
			"topicPath": topicsPath,
		})
		return cache.Topic{}, apperrors.ErrTopicPathNotFound
	}

	return *lastSegmentTopic, nil
}

func getDefaultCensusTopic(ctx context.Context, cacheList cache.List) (cache.Topic, error) {
	censusTopicCache, err := cacheList.CensusTopic.GetCensusData(ctx)
	if err != nil {
		log.Error(ctx, "failed to get census topic cache", err)
		return cache.Topic{}, err
	}
	return *censusTopicCache, nil
}

// Maps template name to underlying go template
func buildDataAggregationPage(w http.ResponseWriter, m model.SearchPage, rend RenderClient, template string) {
	// time-series-tool needs its own template due to the need of elements to be present for JS to be able to assign onClick events(doesn't work if they're conditionally shown on the page)
	if template != "time-series-tool" && template != "search" && template != RelatedPagesTemplate {
		rend.BuildPage(w, m, "data-aggregation-page")
	} else {
		rend.BuildPage(w, m, template)
	}
}

// validateCurrentPage checks if the current page exceeds the total pages which is a bad request
func validateCurrentPage(ctx context.Context, cfg *config.Config, validatedQueryParams data.SearchURLParams, resultsCount int) error {
	if resultsCount > 0 {
		totalPages := data.GetTotalPages(cfg, validatedQueryParams.Limit, resultsCount)

		if validatedQueryParams.CurrentPage > totalPages {
			err := apperrors.ErrPageExceedsTotalPages
			log.Info(ctx, "current page exceeds total pages", log.Data{
				"current_page": validatedQueryParams.CurrentPage,
				"total_pages":  totalPages,
			})

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

	setAuthTokenHeader(options.Headers, accessToken)

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

	log.Info(req.Context(), "setting-response-status", log.Data{
		"status": status,
	})

	if status >= 500 {
		log.Error(req.Context(), "serving internal error status", err)
	}

	w.WriteHeader(status)
}

func setAuthTokenHeader(headers http.Header, accessToken string) {
	if strings.HasPrefix(accessToken, Bearer) {
		headers.Set(searchSDK.Authorization, accessToken)
	} else {
		headers.Set(searchSDK.Authorization, Bearer+accessToken)
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

	setAuthTokenHeader(options.Headers, accessToken)

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
		if len(segments) == 0 { // linter needed this second check
			return nil, fmt.Errorf("no segments to validate")
		}
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
func checkAllowedPageTypes(ctx context.Context, w http.ResponseWriter, zc ZebedeeClient, accessToken, collectionID, lang, pageURL string, allowedPageTypes []string) (zebedeeCli.PageData, error) {
	pageData, err := zc.GetPageData(ctx, accessToken, collectionID, lang, pageURL)
	if err != nil {
		var zebedeeErr zebedeeCli.ErrInvalidZebedeeResponse
		if errors.As(err, &zebedeeErr) {
			errorCode := zebedeeErr.ActualCode
			// Zebedee provides a 400 response if you request a /latest for the wrong type
			// of page. This should be treated as not being able to get the content type.
			if errorCode == 400 || errorCode == 404 {
				log.Info(ctx, "client error getting content type", log.Data{
					"error_code": errorCode,
				})
				return zebedeeCli.PageData{}, apperrors.ErrZebedeePageDataNotFound
			} else {
				log.Error(ctx, "error getting content type", err)
			}
		}
		return zebedeeCli.PageData{}, err
	}

	if !slices.Contains(allowedPageTypes, pageData.Type) {
		log.Info(ctx, "page type isn't compatible with related list page", log.Data{
			"page_type": pageData.Type,
		})
		return zebedeeCli.PageData{}, apperrors.ErrPageTypeIncompatible
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

// postSearchURIs posts a list of URIs to search API and gets a search response
func postSearchURIs(ctx context.Context, searchC SearchClient, options searchSDK.Options, cancel func(), urisRequest searchAPI.URIsRequest) (*searchModels.SearchResponse, searchError.Error, int) {
	if len(urisRequest.URIs) > 0 {
		s, err := searchC.PostSearchURIs(ctx, options, urisRequest)
		if err != nil {
			log.Error(ctx, "getting search response from client failed", err)
			cancel()
			return nil, err, 0
		}
		return s, nil, s.Count
	}
	return nil, nil, 0
}

func selectTopic(ctx context.Context, req *http.Request, cacheList cache.List, aggCfg AggregationConfig) (cache.Topic, error) {
	if aggCfg.TemplateName == RelatedPagesTemplate {
		return cache.Topic{}, nil
	}

	if aggCfg.UseTopicsPath {
		return getSelectedTopic(ctx, req, cacheList)
	}
	return getDefaultCensusTopic(ctx, cacheList)
}

func prepareQueryParams(req *http.Request, aggCfg *AggregationConfig, selectedTopic cache.Topic) bool {
	if aggCfg.URLQueryParams == nil {
		aggCfg.URLQueryParams = req.URL.Query()
	}

	if aggCfg.UseTopicsPath {
		aggCfg.URLQueryParams.Add("topics", selectedTopic.ID)
	}

	// Set the "topics" query parameter to selected topic's query if conditions are met
	if aggCfg.URLQueryParams.Get("topics") == "" && aggCfg.TemplateName == "" {
		aggCfg.URLQueryParams.Add("topics", selectedTopic.Query)
		return true
	}

	return false
}

func validatePageType(ctx context.Context, w http.ResponseWriter, zc ZebedeeClient, aggCfg AggregationConfig, accessToken, collectionID, lang, urlPath string) (zebedeeCli.PageData, error) {
	if aggCfg.TemplateName != RelatedPagesTemplate {
		return zebedeeCli.PageData{}, nil
	}

	validatePath := urlPath
	if !aggCfg.UseURIsRequest {
		validatePath = urlPath + "/latest"
	}

	return checkAllowedPageTypes(ctx, w, zc, accessToken, collectionID, lang, validatePath, knownRelatedListTypes)
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
