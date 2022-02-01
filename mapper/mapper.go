package mapper

import (
	"net/http"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	model "github.com/ONSdigital/dp-frontend-search-controller/model"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// CreateSearchPage maps type searchC.Response to model.Page
func CreateSearchPage(cfg *config.Config, req *http.Request, basePage coreModel.Page, validatedQueryParams data.SearchURLParams, categories []data.Category, respC searchC.Response, departments searchC.Department, lang string) model.SearchPage {

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
	page.URI = req.URL.Path
	page.PatternLibraryAssetsPath = cfg.PatternLibraryAssetsPath
	page.Pagination.CurrentPage = validatedQueryParams.CurrentPage

	mapQuery(cfg, &page, validatedQueryParams, categories, respC)

	mapResponse(&page, respC, categories)

	mapFilters(&page, categories, validatedQueryParams)

	mapDepartments(&page, departments)

	return page
}

func mapQuery(cfg *config.Config, page *model.SearchPage, validatedQueryParams data.SearchURLParams, categories []data.Category, respC searchC.Response) {
	page.Data.Query = validatedQueryParams.Query

	page.Data.Filter = validatedQueryParams.Filter.Query

	mapSort(page, validatedQueryParams)

	mapPagination(cfg, page, validatedQueryParams, respC)
}

func mapSort(page *model.SearchPage, validatedQueryParams data.SearchURLParams) {
	page.Data.Sort.Query = validatedQueryParams.Sort.Query

	page.Data.Sort.LocaliseFilterKeys = validatedQueryParams.Filter.LocaliseKeyName

	page.Data.Sort.LocaliseSortKey = validatedQueryParams.Sort.LocaliseKeyName

	var pageSortOptions []model.SortOptions
	for _, sort := range data.SortOptions {
		pageSortOptions = append(pageSortOptions, model.SortOptions{
			Query:           sort.Query,
			LocaliseKeyName: sort.LocaliseKeyName,
		})
	}
	page.Data.Sort.Options = pageSortOptions
}

func mapPagination(cfg *config.Config, page *model.SearchPage, validatedQueryParams data.SearchURLParams, respC searchC.Response) {
	page.Data.Pagination.Limit = validatedQueryParams.Limit
	page.Data.Pagination.LimitOptions = data.LimitOptions

	page.Data.Pagination.CurrentPage = validatedQueryParams.CurrentPage
	page.Data.Pagination.TotalPages = data.GetTotalPages(cfg, validatedQueryParams.Limit, respC.Count)
	page.Data.Pagination.PagesToDisplay = data.GetPagesToDisplay(cfg, validatedQueryParams, page.Data.Pagination.TotalPages)
	page.Data.Pagination.FirstAndLastPages = data.GetFirstAndLastPages(cfg, validatedQueryParams, page.Data.Pagination.TotalPages)
}

func mapResponse(page *model.SearchPage, respC searchC.Response, categories []data.Category) {
	page.Data.Response.Count = respC.Count

	mapResponseCategories(page, categories)

	mapResponseItems(page, respC)

	page.Data.Response.Suggestions = respC.Suggestions
	page.Data.Response.AdditionalSuggestions = respC.AdditionalSuggestions
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

func mapResponseItems(page *model.SearchPage, respC searchC.Response) {
	itemPage := []model.ContentItem{}

	for _, itemC := range respC.Items {
		item := model.ContentItem{}

		mapItemDescription(&item, itemC)

		mapItemHighlight(&item, itemC)

		item.Type.Type = itemC.Type
		item.Type.LocaliseKeyName = data.GetGroupLocaliseKey(itemC.Type)

		item.URI = itemC.URI

		mapItemMatches(&item, itemC)

		itemPage = append(itemPage, item)
	}

	page.Data.Response.Items = itemPage
}

func mapItemDescription(item *model.ContentItem, itemC searchC.ContentItem) {
	descriptionC := itemC.Description

	item.Description = model.Description{
		DatasetID:         descriptionC.DatasetID,
		Edition:           descriptionC.Edition,
		Headline1:         descriptionC.Headline1,
		Headline2:         descriptionC.Headline2,
		Headline3:         descriptionC.Headline3,
		Keywords:          descriptionC.Keywords,
		LatestRelease:     descriptionC.LatestRelease,
		Language:          descriptionC.Language,
		MetaDescription:   descriptionC.MetaDescription,
		NationalStatistic: descriptionC.NationalStatistic,
		NextRelease:       descriptionC.NextRelease,
		PreUnit:           descriptionC.PreUnit,
		ReleaseDate:       descriptionC.ReleaseDate,
		Source:            descriptionC.Source,
		Summary:           descriptionC.Summary,
		Title:             descriptionC.Title,
		Unit:              descriptionC.Unit,
	}

	if descriptionC.Contact != nil {
		item.Description.Contact = &model.Contact{
			Name:      descriptionC.Contact.Name,
			Telephone: descriptionC.Contact.Telephone,
			Email:     descriptionC.Contact.Email,
		}
	}
}

func mapItemHighlight(item *model.ContentItem, itemC searchC.ContentItem) {
	highlightC := itemC.Description.Highlight

	item.Description.Highlight = model.Highlight{
		DatasetID:       highlightC.DatasetID,
		Edition:         highlightC.Edition,
		Keywords:        highlightC.Keywords,
		MetaDescription: highlightC.MetaDescription,
		Summary:         highlightC.Summary,
		Title:           highlightC.Title,
	}
}

func mapItemMatches(pageItem *model.ContentItem, item searchC.ContentItem) {
	if item.Matches != nil {

		matchesDesc := item.Matches.Description

		pageItem.Matches = &model.Matches{
			Description: model.MatchDescription{},
		}

		// Summary Match
		if matchesDesc.Summary != nil {
			var matchesSummaryPage []model.MatchDetails

			for _, summaryC := range *matchesDesc.Summary {
				matchesSummaryPage = append(matchesSummaryPage, model.MatchDetails{
					Value: summaryC.Value,
					Start: summaryC.Start,
					End:   summaryC.End,
				})
			}

			pageItem.Matches.Description.Summary = &matchesSummaryPage
		}

		// Title Match
		if matchesDesc.Title != nil {
			var matchesTitlePage []model.MatchDetails

			for _, titleC := range *matchesDesc.Title {
				matchesTitlePage = append(matchesTitlePage, model.MatchDetails{
					Value: titleC.Value,
					Start: titleC.Start,
					End:   titleC.End,
				})
			}

			pageItem.Matches.Description.Title = &matchesTitlePage
		}

		// Edition Match
		if matchesDesc.Edition != nil {
			var matchesEditionPage []model.MatchDetails

			for _, editionC := range *matchesDesc.Edition {
				matchesEditionPage = append(matchesEditionPage, model.MatchDetails{
					Value: editionC.Value,
					Start: editionC.Start,
					End:   editionC.End,
				})
			}

			pageItem.Matches.Description.Edition = &matchesEditionPage
		}

		// Meta Description Match
		if matchesDesc.MetaDescription != nil {
			var matchesMetaDescPage []model.MatchDetails

			for _, metaDescC := range *matchesDesc.MetaDescription {
				matchesMetaDescPage = append(matchesMetaDescPage, model.MatchDetails{
					Value: metaDescC.Value,
					Start: metaDescC.Start,
					End:   metaDescC.End,
				})
			}

			pageItem.Matches.Description.MetaDescription = &matchesMetaDescPage
		}

		// Keywords Match
		if matchesDesc.Keywords != nil {
			var matchesKeywordsPage []model.MatchDetails

			for _, keywordC := range *matchesDesc.Keywords {
				matchesKeywordsPage = append(matchesKeywordsPage, model.MatchDetails{
					Value: keywordC.Value,
					Start: keywordC.Start,
					End:   keywordC.End,
				})
			}

			pageItem.Matches.Description.Keywords = &matchesKeywordsPage
		}

		// DatasetID Match
		if matchesDesc.DatasetID != nil {
			var matchesDatasetIDPage []model.MatchDetails

			for _, datasetIDClient := range *matchesDesc.DatasetID {
				matchesDatasetIDPage = append(matchesDatasetIDPage, model.MatchDetails{
					Value: datasetIDClient.Value,
					Start: datasetIDClient.Start,
					End:   datasetIDClient.End,
				})
			}

			pageItem.Matches.Description.DatasetID = &matchesDatasetIDPage
		}
	}
}

func mapFilters(page *model.SearchPage, categories []data.Category, queryParams data.SearchURLParams) {
	var filters []model.Filter

	for _, category := range categories {
		var filter model.Filter
		filter.LocaliseKeyName = category.LocaliseKeyName
		filter.NumberOfResults = category.Count

		var keys []string
		var subTypes []model.Filter
		if len(category.ContentTypes) > 0 {
			for _, contentType := range category.ContentTypes {
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
		filters = append(filters, filter)
	}
	page.Data.Filters = filters
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

func mapDepartments(page *model.SearchPage, departments searchC.Department) {
	if &departments != nil && departments.Items == nil {
		page.Department = nil
		return
	}

	dept := (*departments.Items)[0]
	page.Department = &model.Department{
		Code: dept.Code,
		URL:  dept.URL,
		Name: dept.Name,
	}
	if dept.Matches != nil {
		matches := (*dept.Matches)[0]
		terms := (*matches.Terms)[0]
		page.Department.Match = terms.Value
	}

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
