package handlers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/dp-frontend-search-controller/model"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-topic-api/models"
)

// ReadPreviousReleases handles previous releases page
func (sh *SearchHandler) PreviousReleases(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		previousReleasesConfig := NewPreviousReleasesConfig(*req)

		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, previousReleasesConfig)
	})
}

func NewPreviousReleasesConfig(req http.Request) AggregationConfig {
	createPageModel := func(cfg *config.Config, req *http.Request, base core.Page, queryParams data.SearchURLParams, categories []data.Category, topics []data.Topic, searchResp *searchModels.SearchResponse, lang string, homepageResp zebedee.HomepageContent, errorMessage string, navigationCache *models.Navigation,
		template string, topic cache.Topic, validationErrs []core.ErrorItem, pageData zebedee.PageData, bc []zebedee.Breadcrumb) model.SearchPage {
		return mapper.CreatePreviousReleasesPage(cfg, req, base, queryParams, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
	}
	validateParams := func(ctx context.Context, cfg *config.Config, urlQuery url.Values, urlPath string, _ *cache.Topic) (data.SearchURLParams, []core.ErrorItem) {
		return data.ReviewPreviousReleasesQueryWithParams(ctx, cfg, urlQuery, urlPath)
	}
	getSearchAndCategoriesCountQueries := func(validatedQueryParams data.SearchURLParams, _ *cache.Topic, _, pageDataType string) (searchQuery, categoriesCountQuery url.Values) {
		searchQuery = data.SetParentTypeOnSearchAPIQuery(validatedQueryParams, pageDataType)

		return searchQuery, categoriesCountQuery
	}

	urlQuery := req.URL.Query()
	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)

	return AggregationConfig{
		TemplateName:                       "related-list-pages",
		UseTopicsPath:                      false,
		URLQueryParams:                     sanitisedParams,
		ValidateParams:                     validateParams,
		GetSearchAndCategoriesCountQueries: getSearchAndCategoriesCountQueries,
		CreatePageModel:                    createPageModel,
	}
}
