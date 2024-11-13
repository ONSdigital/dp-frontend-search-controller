package handlers

import (
	"context"
	"net/http"
	"net/url"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
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

// Read Handler
func (sh *SearchHandler) Search(cfg *config.Config, template string) http.HandlerFunc {
	oldHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		searchConfig := NewSearchConfig(false)
		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, searchConfig)
	})

	newHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		searchConfig := NewSearchConfig(true)
		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, searchConfig)
	})

	return cookies.Handler(cfg.ABTest.Enabled, newHandler, oldHandler, cfg.ABTest.Percentage, cfg.ABTest.AspectID, cfg.SiteDomain, cfg.ABTest.Exit)
}

func NewSearchConfig(nlpWeightingEnabled bool) AggregationConfig {
	createPageModel := func(cfg *config.Config, req *http.Request, base core.Page, queryParams data.SearchURLParams, categories []data.Category, topics []data.Topic, searchResp *searchModels.SearchResponse, lang string, homepageResp zebedeeCli.HomepageContent, errorMessage string, navigationCache *models.Navigation,
		template string, topic cache.Topic, validationErrs []core.ErrorItem, pageData zebedeeCli.PageData, _ []zebedeeCli.Breadcrumb) model.SearchPage {
		return mapper.CreateSearchPage(cfg, req, base, queryParams, categories, topics, searchResp, lang, homepageResp, errorMessage, navigationCache, validationErrs)
	}
	validateParams := func(ctx context.Context, cfg *config.Config, urlQuery url.Values, _ string, censusTopicCache *cache.Topic) (data.SearchURLParams, []core.ErrorItem) {
		return data.ReviewQuery(ctx, cfg, urlQuery, censusTopicCache)
	}
	getSearchAndCategoriesCountQueries := func(validatedQueryParams data.SearchURLParams, censusTopicCache *cache.Topic, _, _ string) (searchQuery, categoriesCountQuery url.Values) {
		searchQuery = data.GetSearchAPIQuery(validatedQueryParams, censusTopicCache)
		categoriesCountQuery = getCategoriesCountQuery(searchQuery)

		return searchQuery, categoriesCountQuery
	}

	return AggregationConfig{
		TemplateName:                       "search",
		UseTopicsPath:                      false,
		ValidateParams:                     validateParams,
		GetSearchAndCategoriesCountQueries: getSearchAndCategoriesCountQueries,
		CreatePageModel:                    createPageModel,
		NLPWeightingEnabled:                nlpWeightingEnabled,
	}
}
