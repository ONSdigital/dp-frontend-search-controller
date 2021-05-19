package mapper

import (
	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	model "github.com/ONSdigital/dp-frontend-models/model/search"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
)

// CreateSearchPage maps type searchC.Response to model.Page
func CreateSearchPage(cfg *config.Config, validatedQueryParams data.SearchURLParams, categories []data.Category, respC searchC.Response, lang string) (page model.Page) {
	// SEARCH STRUCT MAPPING
	page.Metadata.Title = "Search"
	page.SearchDisabled = true
	page.Language = lang

	mapQuery(cfg, &page, validatedQueryParams, categories, respC)

	mapResponse(&page, respC, categories)

	return page
}

func mapQuery(cfg *config.Config, page *model.Page, validatedQueryParams data.SearchURLParams, categories []data.Category, respC searchC.Response) {
	page.Data.Query = validatedQueryParams.Query

	page.Data.Filter = validatedQueryParams.Filter.Query

	mapSort(page, validatedQueryParams)

	mapPagination(cfg, page, validatedQueryParams, respC)
}

func mapSort(page *model.Page, validatedQueryParams data.SearchURLParams) {
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

func mapPagination(cfg *config.Config, page *model.Page, validatedQueryParams data.SearchURLParams, respC searchC.Response) {
	page.Data.Pagination.Limit = validatedQueryParams.Limit
	page.Data.Pagination.LimitOptions = data.LimitOptions

	page.Data.Pagination.CurrentPage = validatedQueryParams.CurrentPage
	page.Data.Pagination.TotalPages = data.GetTotalPages(validatedQueryParams.Limit, respC.Count)
	page.Data.Pagination.PagesToDisplay = data.GetPagesToDisplay(cfg, validatedQueryParams, page.Data.Pagination.TotalPages)
}

func mapResponse(page *model.Page, respC searchC.Response, categories []data.Category) {
	page.Data.Response.Count = respC.Count

	mapResponseCategories(page, categories)

	mapResponseItems(page, respC)

	page.Data.Response.Suggestions = respC.Suggestions
}

func mapResponseCategories(page *model.Page, categories []data.Category) {
	pageCategories := []model.Category{}

	for _, category := range categories {
		pageContentType := []model.ContentType{}

		for _, contentType := range category.ContentTypes {
			pageContentType = append(pageContentType, model.ContentType{
				Type:            contentType.Type,
				Count:           contentType.Count,
				LocaliseKeyName: contentType.LocaliseKeyName,
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

func mapResponseItems(page *model.Page, respC searchC.Response) {
	itemPage := []model.ContentItem{}

	for _, itemC := range respC.Items {
		item := model.ContentItem{}

		mapItemDescription(&item, itemC)

		item.Type = itemC.Type

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

func mapItemMatches(item *model.ContentItem, itemC searchC.ContentItem) {
	if itemC.Matches != nil {

		matchesDescC := itemC.Matches.Description

		item.Matches = &model.Matches{
			Description: model.MatchDescription{},
		}

		// Summary Match
		if matchesDescC.Summary != nil {
			var matchesSummaryPage []model.MatchDetails

			for _, summaryC := range *matchesDescC.Summary {
				matchesSummaryPage = append(matchesSummaryPage, model.MatchDetails{
					Value: summaryC.Value,
					Start: summaryC.Start,
					End:   summaryC.End,
				})
			}

			item.Matches.Description.Summary = &matchesSummaryPage
		}

		// Title Match
		if matchesDescC.Title != nil {
			var matchesTitlePage []model.MatchDetails

			for _, titleC := range *matchesDescC.Title {
				matchesTitlePage = append(matchesTitlePage, model.MatchDetails{
					Value: titleC.Value,
					Start: titleC.Start,
					End:   titleC.End,
				})
			}

			item.Matches.Description.Title = &matchesTitlePage
		}

		// Edition Match
		if matchesDescC.Edition != nil {
			var matchesEditionPage []model.MatchDetails

			for _, editionC := range *matchesDescC.Edition {
				matchesEditionPage = append(matchesEditionPage, model.MatchDetails{
					Value: editionC.Value,
					Start: editionC.Start,
					End:   editionC.End,
				})
			}

			item.Matches.Description.Edition = &matchesEditionPage
		}

		// Meta Description Match
		if matchesDescC.MetaDescription != nil {
			var matchesMetaDescPage []model.MatchDetails

			for _, metaDescC := range *matchesDescC.MetaDescription {
				matchesMetaDescPage = append(matchesMetaDescPage, model.MatchDetails{
					Value: metaDescC.Value,
					Start: metaDescC.Start,
					End:   metaDescC.End,
				})
			}

			item.Matches.Description.MetaDescription = &matchesMetaDescPage
		}

		// Keywords Match
		if matchesDescC.Keywords != nil {
			var matchesKeywordsPage []model.MatchDetails

			for _, keywordC := range *matchesDescC.Keywords {
				matchesKeywordsPage = append(matchesKeywordsPage, model.MatchDetails{
					Value: keywordC.Value,
					Start: keywordC.Start,
					End:   keywordC.End,
				})
			}

			item.Matches.Description.Keywords = &matchesKeywordsPage
		}

		// DatasetID Match
		if matchesDescC.DatasetID != nil {
			var matchesDatasetIDPage []model.MatchDetails

			for _, datasetIDClient := range *matchesDescC.DatasetID {
				matchesDatasetIDPage = append(matchesDatasetIDPage, model.MatchDetails{
					Value: datasetIDClient.Value,
					Start: datasetIDClient.Start,
					End:   datasetIDClient.End,
				})
			}

			item.Matches.Description.DatasetID = &matchesDatasetIDPage
		}
	}
}
