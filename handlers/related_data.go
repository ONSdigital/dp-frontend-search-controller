package handlers

import (
	"context"
	"net/http"
	"path"
	"sync"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	searchAPI "github.com/ONSdigital/dp-search-api/api"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// list of content types that have /previousreleases
var knownRelatedDataTypesTypes = []string{
	"bulletin",
	"article",
	"compendium_landing_page",
}

func readRelatedData(w http.ResponseWriter, req *http.Request, cfg *config.Config, zc ZebedeeClient, rend RenderClient, searchC SearchClient,
	accessToken, collectionID, lang string, cacheList cache.List,
) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	var err error
	template := "related-list-pages"
	urlPath := path.Dir(req.URL.Path)
	urlQuery := req.URL.Query()

	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)

	// check page type
	pageData, err := checkAllowedPageTypes(ctx, w, zc, accessToken, collectionID, lang, urlPath, knownRelatedDataTypesTypes)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get cached navigation data
	navigationCache := getNavigationCache(ctx, w, req, cacheList, lang)

	validatedQueryParams, validationErrs := data.ReviewPreviousReleasesQueryWithParams(ctx, cfg, sanitisedParams, urlPath)

	if len(validationErrs) > 0 {
		log.Info(ctx, "validation of parameters failed", log.Data{
			"parameters": validationErrs,
		})
		// Errors are now mapped to the page model to output feedback to the user rather than
		// a blank 400 error response.
		m := mapper.CreateRelatedDataPage(cfg, req, rend.NewBasePageModel(), validatedQueryParams, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs, zebedeeCli.PageData{}, []zebedeeCli.Breadcrumb{})
		rend.BuildPage(w, m, template)
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

		searchRespErr error
		searchCount   int
	)
	wg.Add(counter)

	go func() {
		defer wg.Done()
		homepageResp = getHomepageContent(ctx, zc, accessToken, collectionID, lang)
	}()

	var options searchSDK.Options
	options.Query = searchQuery
	options.Headers = http.Header{
		searchSDK.FlorenceToken: {Bearer + accessToken},
		searchSDK.CollectionID:  {collectionID},
	}

	var URIList []string
	for _, related := range pageData.RelatedData {
		URIList = append(URIList, related.URI)
	}

	URIsRequest := searchAPI.URIsRequest{
		URIs:   URIList,
		Limit:  validatedQueryParams.Limit,
		Offset: validatedQueryParams.Offset,
	}

	go func() {
		defer wg.Done()
		searchResp, searchRespErr, searchCount = postSearchURIs(ctx, searchC, options, cancel, URIsRequest)
	}()

	go func() {
		defer wg.Done()
		bc = getBreadcrumb(ctx, zc, accessToken, collectionID, lang, urlPath)
	}()

	wg.Wait()
	if searchRespErr != nil {
		setStatusCode(w, req, searchRespErr)
		return
	}

	basePage := rend.NewBasePageModel()
	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchCount)
	if err != nil {
		validationErrs = append(validationErrs, core.ErrorItem{
			Description: core.Localisation{
				Text: "current page exceeds total pages",
			},
		})
		m := mapper.CreateRelatedDataPage(cfg, req, basePage, validatedQueryParams, &searchModels.SearchResponse{}, lang, zebedeeCli.HomepageContent{}, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
		rend.BuildPage(w, m, template)
		return
	}

	m := mapper.CreateRelatedDataPage(cfg, req, basePage, validatedQueryParams, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
	rend.BuildPage(w, m, template)
}
