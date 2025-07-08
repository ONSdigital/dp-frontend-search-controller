package handlers

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/dp-frontend-search-controller/model"
	dphandlers "github.com/ONSdigital/dp-net/v3/handlers"
	core "github.com/ONSdigital/dp-renderer/v2/model"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-topic-api/models"
)

// RelatedData handler
func (sh *SearchHandler) RelatedData(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		relatedDataConfig := NewRelatedDataConfig(*req)
		ctx := context.Background()
		urlPath := path.Dir(req.URL.Path)

		if strings.HasSuffix(urlPath, "/latest") {
			previousRelease, err := validatePageType(ctx, w, sh.ZebedeeClient, relatedDataConfig, accessToken, collectionID, lang, urlPath)
			if err != nil {
				setStatusCode(w, req, err)
				return
			}

			if previousRelease.Description.MigrationLink != "" {
				http.Redirect(w, req, previousRelease.Description.MigrationLink+"/related-data", http.StatusPermanentRedirect)
				return
			}
		}

		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, relatedDataConfig)
	})
}

func NewRelatedDataConfig(req http.Request) AggregationConfig {
	createPageModel := func(cfg *config.Config, req *http.Request, base core.Page, queryParams data.SearchURLParams, categories []data.Category, topics []data.Topic, searchResp *searchModels.SearchResponse, lang string, homepageResp zebedeeCli.HomepageContent, errorMessage string, navigationCache *models.Navigation,
		template string, topic cache.Topic, validationErrs []core.ErrorItem, pageData zebedeeCli.PageData, bc []zebedeeCli.Breadcrumb) model.SearchPage {
		return mapper.CreateRelatedDataPage(cfg, req, base, queryParams, searchResp, lang, homepageResp, "", navigationCache, template, cache.Topic{}, validationErrs, pageData, bc)
	}
	validateParams := func(ctx context.Context, cfg *config.Config, urlQuery url.Values, urlPath string, _ *cache.Topic) (data.SearchURLParams, []core.ErrorItem) {
		return data.ReviewRelatedDataQueryWithParams(ctx, cfg, urlQuery, urlPath)
	}
	getSearchAndCategoriesCountQueries := func(validatedQueryParams data.SearchURLParams, _ *cache.Topic, parentType, _ string) (searchQuery, categoriesCountQuery url.Values) {
		searchQuery = data.SetParentTypeOnSearchAPIQuery(validatedQueryParams, parentType)

		return searchQuery, categoriesCountQuery
	}

	urlQuery := req.URL.Query()
	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)

	return AggregationConfig{
		TemplateName:                       "related-list-pages",
		UseTopicsPath:                      false,
		UseURIsRequest:                     true,
		URLQueryParams:                     sanitisedParams,
		ValidateParams:                     validateParams,
		GetSearchAndCategoriesCountQueries: getSearchAndCategoriesCountQueries,
		CreatePageModel:                    createPageModel,
	}
}
