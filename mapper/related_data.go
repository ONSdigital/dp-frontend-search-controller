package mapper

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/model"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	topicModel "github.com/ONSdigital/dp-topic-api/models"
)

func CreateRelatedDataPage(cfg *config.Config, req *http.Request, basePage coreModel.Page,
	validatedQueryParams data.SearchURLParams, respC *searchModels.SearchResponse, lang string, homepageResponse zebedee.HomepageContent, errorMessage string,
	navigationContent *topicModel.Navigation, template string, topic cache.Topic, validationErrs []coreModel.ErrorItem, zebedeeResp zebedee.PageData, bc []zebedee.Breadcrumb,
) model.SearchPage {
	page := model.SearchPage{
		Page: basePage,
	}

	page.Metadata.Title = "Data related to " + zebedeeResp.Description.Title
	page.Title.Title = "All data related to " + zebedeeResp.Description.Title + ": " + zebedeeResp.Description.Edition
	page.Metadata.Description = zebedeeResp.Description.MetaDescription
	page.Type = "related-data"
	page.Language = lang
	page.BetaBannerEnabled = true
	page.SearchDisabled = false
	page.Pagination.CurrentPage = validatedQueryParams.CurrentPage
	page.ServiceMessage = homepageResponse.ServiceMessage
	page.EmergencyBanner = mapEmergencyBanner(homepageResponse)

	if respC != nil {
		page.Count = respC.Count
	} else {
		page.Count = 0
	}

	MapCookiePreferences(req, &page.Page.CookiesPreferencesSet, &page.Page.CookiesPolicy)

	mapQuery(cfg, &page, validatedQueryParams, respC, *req, errorMessage)

	mapResponse(&page, respC, []data.Category{})

	mapBreadcrumb(&page, bc, zebedeeResp.Description.Title, zebedeeResp.URI)
	return page
}
