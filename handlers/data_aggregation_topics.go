package handlers

import (
	"context"
	"net/http"
	"net/url"

	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
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

// ReadDataAggregationWithTopics for data aggregation routes with topic/subtopics
func (sh *SearchHandler) DataAggregationWithTopics(cfg *config.Config, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		aggregationConfig := NewAggregationWithTopicsConfig(template)

		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, aggregationConfig)
	})
}

func NewAggregationWithTopicsConfig(template string) AggregationConfig {
	createPageModel := func(cfg *config.Config, req *http.Request, base core.Page, queryParams data.SearchURLParams, categories []data.Category, topics []data.Topic, searchResp *searchModels.SearchResponse, lang string, homepageResp zebedeeCli.HomepageContent, errorMessage string, navigationCache *models.Navigation,
		template string, topic cache.Topic, validationErrs []core.ErrorItem, _ zebedeeCli.PageData, _ []zebedeeCli.Breadcrumb) model.SearchPage {
		return mapper.CreateDataAggregationPage(cfg, req, base, queryParams, categories, topics, searchResp, lang, homepageResp, errorMessage, navigationCache, template, topic, validationErrs)
	}
	validateParams := func(ctx context.Context, cfg *config.Config, urlQuery url.Values, _ string, _ *cache.Topic) (data.SearchURLParams, []core.ErrorItem) {
		return data.ReviewDataAggregationQueryWithParams(ctx, cfg, urlQuery)
	}

	getSearchAndCategoriesCountQueries := func(validatedQueryParams data.SearchURLParams, _ *cache.Topic, template, _ string) (searchQuery, categoriesCountQuery url.Values) {
		searchQuery = data.GetDataAggregationQuery(validatedQueryParams, template)
		categoriesCountQuery = getCategoriesTopicsCountQuery(searchQuery)

		return searchQuery, categoriesCountQuery
	}

	return AggregationConfig{
		TemplateName:                       template,
		UseTopicsPath:                      true,
		ValidateParams:                     validateParams,
		GetSearchAndCategoriesCountQueries: getSearchAndCategoriesCountQueries,
		CreatePageModel:                    createPageModel,
	}
}
