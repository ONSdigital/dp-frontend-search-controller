package handlers

import (
	"context"
	"net/http"
	"path"
	"sync"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	searchAPI "github.com/ONSdigital/dp-search-api/api"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// list of content types that have /relateddata
var knownRelatedDataTypes = []string{
	"bulletin",
	"article",
	"compendium_landing_page",
}

// ReadRelated data handles related data page
func (sh *SearchHandler) RelatedData(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readRelatedData(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList)
	})
}

func readRelatedData(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	template := "related-list-pages"
	urlPath := path.Dir(req.URL.Path)
	urlQuery := req.URL.Query()

	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)

	// check page type
	pageData, err := checkAllowedPageTypes(ctx, w, zc, accessToken, collectionID, lang, urlPath, knownRelatedDataTypes)
	if err != nil {
		log.Error(ctx, "page type isn't compatible with /relateddata", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// counter used to keep track of the number of concurrent API calls
	var counter = 3
	var (
		navigationCache *models.Navigation
		bc              []zebedeeCli.Breadcrumb
		homepageResp    zebedeeCli.HomepageContent

		wg sync.WaitGroup
	)

	wg.Add(counter)

	go func() {
		defer wg.Done()
		// get cached navigation data
		navigationCache = getNavigationCache(ctx, w, req, cacheList, lang)
	}()

	go func() {
		defer wg.Done()
		// get breadcrumbs
		bc = getBreadcrumb(ctx, zc, accessToken, collectionID, lang, urlPath)
	}()

	go func() {
		defer wg.Done()
		// get homepage content
		homepageResp = getHomepageContent(ctx, zc, accessToken, collectionID, lang)
	}()

	wg.Wait()

	validatedQueryParams, validationErrs := data.ReviewPreviousReleasesQueryWithParams(ctx, cfg, sanitisedParams, urlPath)

	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed", log.Data{
			"parameters": validationErrs,
		})
		// Errors are now mapped to the page model to output feedback to the user rather than
		// a blank 400 error response.
		m := mapper.CreateRelatedDataPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, &searchModels.SearchResponse{}, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
		rend.BuildPage(w, m, template)
		return
	}

	searchQuery := data.SetParentTypeOnSearchAPIQuery(validatedQueryParams, pageData.Type)

	var options searchSDK.Options
	options.Query = searchQuery
	options.Headers = http.Header{
		searchSDK.CollectionID: {collectionID},
	}

	setAuthTokenHeader(options.Headers, accessToken)

	URIList := make([]string, 0, len(pageData.RelatedData))
	for _, related := range pageData.RelatedData {
		URIList = append(URIList, related.URI)
	}

	URIsRequest := searchAPI.URIsRequest{
		URIs:   URIList,
		Limit:  validatedQueryParams.Limit,
		Offset: validatedQueryParams.Offset,
	}

	searchResp, searchRespErr, searchCount := postSearchURIs(ctx, searchC, options, cancel, URIsRequest)
	if searchRespErr != nil {
		setStatusCode(w, req, searchRespErr)
		return
	}

	pErr := validateCurrentPage(ctx, cfg, validatedQueryParams, searchCount)
	if pErr != nil {
		log.Info(ctx, apperrors.ErrPageExceedsTotalPages.Error())
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: apperrors.ErrPageExceedsTotalPages.Error(),
			},
		})
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateRelatedDataPage(cfg, req, basePage, validatedQueryParams, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
	rend.BuildPage(w, m, template)
}
