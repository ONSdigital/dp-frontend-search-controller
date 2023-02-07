package mapper

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"

	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	model "github.com/ONSdigital/dp-frontend-search-controller/model"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	topicModel "github.com/ONSdigital/dp-topic-api/models"
)

// CreateSearchPage maps type searchC.Response to model.Page
func CreateSearchPage(cfg *config.Config, req *http.Request, basePage coreModel.Page,
	validatedQueryParams data.SearchURLParams, categories []data.Category, topicCategories []data.Topic,
	respC *searchModels.SearchResponse, lang string, homepageResponse zebedee.HomepageContent, errorMessage string,
	navigationContent *topicModel.Navigation) model.SearchPage {
	page := model.SearchPage{
		Page: basePage,
	}

	MapCookiePreferences(req, &page.Page.CookiesPreferencesSet, &page.Page.CookiesPolicy)

	page.Metadata.Title = "Search"
	page.Type = "search"
	page.Count = respC.Count
	page.Language = lang
	page.BetaBannerEnabled = true
	page.SearchDisabled = false
	page.URI = req.URL.RequestURI()
	page.PatternLibraryAssetsPath = cfg.PatternLibraryAssetsPath
	page.Pagination.CurrentPage = validatedQueryParams.CurrentPage
	page.ServiceMessage = homepageResponse.ServiceMessage
	page.EmergencyBanner = mapEmergencyBanner(homepageResponse)
	page.SearchNoIndexEnabled = cfg.NoIndexEnabled
	if navigationContent != nil {
		page.NavigationContent = mapNavigationContent(*navigationContent)
	}

	mapQuery(cfg, &page, validatedQueryParams, categories, respC, errorMessage)

	mapResponse(&page, respC, categories)

	mapFilters(&page, categories, validatedQueryParams)

	mapTopicFilters(cfg, &page, topicCategories, validatedQueryParams)

	return page
}

func mapQuery(cfg *config.Config, page *model.SearchPage, validatedQueryParams data.SearchURLParams, categories []data.Category, respC *searchModels.SearchResponse, errorMessage string) {
	page.Data.Query = validatedQueryParams.Query

	page.Data.Filter = validatedQueryParams.Filter.Query

	page.Data.ErrorMessage = errorMessage

	mapSort(page, validatedQueryParams)

	mapPagination(cfg, page, validatedQueryParams, respC)
}

func mapSort(page *model.SearchPage, validatedQueryParams data.SearchURLParams) {
	page.Data.Sort.Query = validatedQueryParams.Sort.Query

	page.Data.Sort.LocaliseFilterKeys = validatedQueryParams.Filter.LocaliseKeyName

	page.Data.Sort.LocaliseSortKey = validatedQueryParams.Sort.LocaliseKeyName

	pageSortOptions := make([]model.SortOptions, len(data.SortOptions))
	for i := range data.SortOptions {
		pageSortOptions[i] = model.SortOptions{
			Query:           data.SortOptions[i].Query,
			LocaliseKeyName: data.SortOptions[i].LocaliseKeyName,
		}
	}

	page.Data.Sort.Options = pageSortOptions
}

func mapPagination(cfg *config.Config, page *model.SearchPage, validatedQueryParams data.SearchURLParams, respC *searchModels.SearchResponse) {
	page.Data.Pagination.Limit = validatedQueryParams.Limit
	page.Data.Pagination.LimitOptions = data.LimitOptions

	page.Data.Pagination.CurrentPage = validatedQueryParams.CurrentPage
	page.Data.Pagination.TotalPages = data.GetTotalPages(cfg, validatedQueryParams.Limit, respC.Count)
	page.Data.Pagination.PagesToDisplay = data.GetPagesToDisplay(cfg, validatedQueryParams, page.Data.Pagination.TotalPages)
	page.Data.Pagination.FirstAndLastPages = data.GetFirstAndLastPages(cfg, validatedQueryParams, page.Data.Pagination.TotalPages)
}

func mapResponse(page *model.SearchPage, respC *searchModels.SearchResponse, categories []data.Category) {
	page.Data.Response.Count = respC.Count

	mapResponseCategories(page, categories)

	mapResponseItems(page, respC)

	page.Data.Response.Suggestions = respC.Suggestions
	page.Data.Response.AdditionalSuggestions = respC.AdditionSuggestions
}

func mapResponseItems(page *model.SearchPage, respC *searchModels.SearchResponse) {
	itemPage := []model.ContentItem{}
	for i := range respC.Items {
		item := model.ContentItem{}

		mapItemDescription(&item, &respC.Items[i])

		mapItemHighlight(&item, &respC.Items[i])

		item.Type.Type = respC.Items[i].DataType
		item.Type.LocaliseKeyName = data.GetGroupLocaliseKey(respC.Items[i].DataType)

		item.URI = respC.Items[i].URI

		itemPage = append(itemPage, item)
	}

	page.Data.Response.Items = itemPage
}

func mapItemDescription(item *model.ContentItem, itemC *searchModels.ESSourceDocument) {
	item.Description = model.Description{
		DatasetID:       itemC.DatasetID,
		Language:        itemC.Language,
		MetaDescription: itemC.MetaDescription,
		ReleaseDate:     itemC.ReleaseDate,
		Summary:         itemC.Summary,
		Title:           itemC.Title,
	}

	if len(itemC.Keywords) != 0 {
		item.Description.Keywords = itemC.Keywords
	} else {
		item.Description.Keywords = nil
	}
}

func mapItemHighlight(item *model.ContentItem, itemC *searchModels.ESSourceDocument) {
	itemHighlight := itemC.Highlight
	if !reflect.ValueOf(itemHighlight).IsNil() {
		item.Description.Highlight = model.Highlight{
			DatasetID:       itemHighlight.DatasetID,
			Edition:         itemC.Edition,
			Keywords:        itemHighlight.Keywords,
			MetaDescription: itemHighlight.MetaDescription,
			Summary:         itemHighlight.Summary,
			Title:           itemHighlight.Title,
		}
	} else {
		item.Description.Highlight = model.Highlight{}
	}
}

func mapResponseCategories(page *model.SearchPage, categories []data.Category) {
	pageCategories := []model.Category{}

	for _, category := range categories {
		pageContentType := []model.ContentType{}

		for _, contentType := range category.ContentTypes {
			pageContentType = append(pageContentType, model.ContentType{
				Group:           contentType.Group,
				Count:           contentType.Count,
				LocaliseKeyName: contentType.LocaliseKeyName,
				Types:           contentType.Types,
			})
		}

		pageCategories = append(pageCategories, model.Category{
			Count:           category.Count,
			LocaliseKeyName: category.LocaliseKeyName,
			ContentTypes:    pageContentType,
		})
	}

	page.Data.Response.Categories = pageCategories
}

func mapFilters(page *model.SearchPage, categories []data.Category, queryParams data.SearchURLParams) {
	filters := make([]model.Filter, len(categories))

	for i := range categories {
		var filter model.Filter
		filter.LocaliseKeyName = categories[i].LocaliseKeyName
		filter.NumberOfResults = categories[i].Count

		var keys []string
		var subTypes []model.Filter
		if len(categories[i].ContentTypes) > 0 {
			for _, contentType := range categories[i].ContentTypes {
				if !contentType.ShowInWebUI && contentType.Count > 0 {
					filter.NumberOfResults -= contentType.Count
					continue
				}
				var subType model.Filter
				subType.LocaliseKeyName = contentType.LocaliseKeyName
				subType.NumberOfResults = contentType.Count
				subType.FilterKey = []string{contentType.Group}

				isChecked := mapIsChecked(subType.FilterKey, queryParams)
				subType.IsChecked = isChecked
				subTypes = append(subTypes, subType)

				keys = append(keys, contentType.Group)
			}
		}

		filter.Types = subTypes
		filter.FilterKey = keys
		filter.IsChecked = mapIsChecked(filter.FilterKey, queryParams)
		filters[i] = filter
	}

	page.Data.Filters = filters
}

func mapTopicFilters(cfg *config.Config, page *model.SearchPage, topicCategories []data.Topic, queryParams data.SearchURLParams) {
	if !cfg.EnableCensusTopicFilterOption {
		return
	}

	var topicsQueryParam []string
	if queryParams.TopicFilter != "" {
		topicsQueryParam = strings.Split(queryParams.TopicFilter, ",")
	}

	mapTopicQueryParams := make(map[string]bool)
	for i := range topicsQueryParam {
		mapTopicQueryParams[topicsQueryParam[i]] = true
	}

	topicFilters := make([]model.TopicFilter, len(topicCategories))

	for i := range topicCategories {
		if !topicCategories[i].ShowInWebUI {
			continue
		}

		var topicFilter model.TopicFilter

		topicFilter.LocaliseKeyName = topicCategories[i].LocaliseKeyName
		topicFilter.NumberOfResults = topicCategories[i].Count
		topicFilter.Query = topicCategories[i].Query
		topicFilter.DistinctItemsCount = topicCategories[i].DistinctItemsCount

		if len(topicsQueryParam) > 0 {
			topicFilter.IsChecked = true
		}

		topicFilters[i] = topicFilter

		for j := range topicCategories[i].Subtopics {
			if !topicCategories[i].Subtopics[j].ShowInWebUI {
				continue
			}
			var subtopicFilter model.TopicFilter

			subtopicFilter.LocaliseKeyName = topicCategories[i].Subtopics[j].LocaliseKeyName
			subtopicFilter.NumberOfResults = topicCategories[i].Subtopics[j].Count
			subtopicFilter.Query = topicCategories[i].Subtopics[j].Query

			if mapTopicQueryParams[topicCategories[i].Subtopics[j].Query] {
				subtopicFilter.IsChecked = true
			}

			topicFilters[i].Types = append(topicFilters[i].Types, subtopicFilter)
		}
	}

	page.Data.TopicFilters = topicFilters
}

func mapIsChecked(contentTypes []string, queryParams data.SearchURLParams) bool {
	for _, query := range queryParams.Filter.Query {
		for _, contentType := range contentTypes {
			if query == contentType {
				return true
			}
		}
	}

	return false
}

// MapCookiePreferences reads cookie policy and preferences cookies and then maps the values to the page model
func MapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *coreModel.CookiesPolicy) {
	preferencesCookie := cookies.GetCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet

	*policy = coreModel.CookiesPolicy{
		Essential: preferencesCookie.Policy.Essential,
		Usage:     preferencesCookie.Policy.Usage,
	}
}

func mapEmergencyBanner(hpc zebedee.HomepageContent) coreModel.EmergencyBanner {
	var mappedEmergencyBanner coreModel.EmergencyBanner
	emptyBannerObj := zebedee.EmergencyBanner{}
	bannerData := hpc.EmergencyBanner

	if bannerData != emptyBannerObj {
		mappedEmergencyBanner.Title = bannerData.Title
		mappedEmergencyBanner.Type = strings.Replace(bannerData.Type, "_", "-", -1)
		mappedEmergencyBanner.Description = bannerData.Description
		mappedEmergencyBanner.URI = bannerData.URI
		mappedEmergencyBanner.LinkText = bannerData.LinkText
	}

	return mappedEmergencyBanner
}

// mapNavigationContent takes navigationContent as returned from the client and returns information needed for the navigation bar
func mapNavigationContent(navigationContent topicModel.Navigation) []coreModel.NavigationItem {
	var mappedNavigationContent []coreModel.NavigationItem

	if navigationContent.Items != nil {
		for _, rootContent := range *navigationContent.Items {
			var subItems []coreModel.NavigationItem

			if rootContent.SubtopicItems != nil {
				for _, subtopicContent := range *rootContent.SubtopicItems {
					subItems = append(subItems, coreModel.NavigationItem{
						Uri:   subtopicContent.Uri,
						Label: subtopicContent.Label,
					})
				}
			}

			mappedNavigationContent = append(mappedNavigationContent, coreModel.NavigationItem{
				Uri:      rootContent.Uri,
				Label:    rootContent.Label,
				SubItems: subItems,
			})
		}
	}

	return mappedNavigationContent
}
