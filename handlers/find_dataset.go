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

func (sh *SearchHandler) FindDataset(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		findDatasetConfig := NewFindDatasetConfig(req)

		handleReadRequest(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, findDatasetConfig)
	})
}

func NewFindDatasetConfig(req *http.Request) AggregationConfig {
	createPageModel := func(cfg *config.Config, req *http.Request, basePage core.Page, validatedQueryParams data.SearchURLParams, categories []data.Category, topicCategories []data.Topic, searchResp *searchModels.SearchResponse, lang string, homepageResp zebedee.HomepageContent, _ string, navigationCache *models.Navigation, _ string, censusTopicCache cache.Topic, errorMessage []core.ErrorItem, _ zebedee.PageData, _ []zebedee.Breadcrumb) model.SearchPage {
		return mapper.CreateDataFinderPage(cfg, req, basePage, validatedQueryParams, categories, topicCategories, []data.PopulationTypes{}, []data.Dimensions{}, searchResp, lang, homepageResp, errorMessage, navigationCache)
	}
	validateParams := func(ctx context.Context, cfg *config.Config, urlQuery url.Values, _ string, censusTopicCache *cache.Topic) (data.SearchURLParams, []core.ErrorItem) {
		return data.ReviewDatasetQuery(ctx, cfg, urlQuery, censusTopicCache)
	}

	getSearchAndCategoriesCountQueries := func(validatedQueryParams data.SearchURLParams, censusTopicCache *cache.Topic, _, _ string) (searchQuery, categoriesCountQuery url.Values) {
		searchQuery = data.GetSearchAPIQuery(validatedQueryParams, censusTopicCache)
		categoriesCountQuery = getCategoriesDatasetCountQuery(searchQuery)

		return searchQuery, categoriesCountQuery
	}

	urlQuery := req.URL.Query()
	urlQuery.Del("filter")
	urlQuery.Add("filter", "dataset_landing_page")
	urlQuery.Add("filter", "user_requested_data")

	sanitisedParams := sanitiseQueryParams(allowedPreviousReleasesQueryParams, urlQuery)

	return AggregationConfig{
		TemplateName:                       "",
		UseTopicsPath:                      false,
		URLQueryParams:                     sanitisedParams,
		ValidateParams:                     validateParams,
		GetSearchAndCategoriesCountQueries: getSearchAndCategoriesCountQueries,
		CreatePageModel:                    createPageModel,
	}
}
