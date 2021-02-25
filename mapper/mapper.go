package mapper

import (
	"context"
	"net/url"
	"strconv"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
)

// CreateSearchPage maps type search.Response to model.Page
func CreateSearchPage(ctx context.Context, url *url.URL, respC searchC.Response, categories []data.Category, paginationQuery *data.PaginationQuery) (page search.Page) {
	// SEARCH STRUCT MAPPING
	query := url.Query()
	page.Metadata.Title = "Search"
	page.SearchDisabled = true
	page.Data.Query = query.Get("q")
	page.Data.Filter = query["filter"]
	page.Data.Sort.Query = query.Get("sort")
	page.Data.Sort.LocaliseFilterKeys = getFilterSortKeyList(query, categories)
	page.Data.Sort.LocaliseSortKey = getSortLocaliseKey(query)

	pageSortOptions := []search.SortOptions{}
	for _, sort := range data.SortOptions {
		pageSortOptions = append(pageSortOptions, search.SortOptions{
			Query:           sort.Query,
			LocaliseKeyName: sort.LocaliseKeyName,
		})
	}
	page.Data.Sort.Options = pageSortOptions

	page.Data.Pagination.LimitOptions = data.GetLimitOptions()
	page.Data.Pagination.Limit = paginationQuery.Limit
	page.Data.Pagination.TotalPages = (respC.Count + page.Data.Pagination.Limit - 1) / page.Data.Pagination.Limit
	page.Data.Pagination.CurrentPage = paginationQuery.CurrentPage
	page.Data.Pagination.PagesToDisplay = getPagesToDisplay(page.Data.Pagination.CurrentPage, page.Data.Pagination.TotalPages, url)

	//RESPONSE STRUCT MAPPING
	page.Data.Response.Count = respC.Count

	pageCategories := []search.Category{}
	for _, category := range categories {
		pageContentType := []search.ContentType{}
		for _, contentType := range category.ContentTypes {
			pageContentType = append(pageContentType, search.ContentType{
				Type:            contentType.Type,
				Count:           contentType.Count,
				LocaliseKeyName: contentType.LocaliseKeyName,
			})
		}
		pageCategories = append(pageCategories, search.Category{
			Count:           category.Count,
			LocaliseKeyName: category.LocaliseKeyName,
			ContentTypes:    pageContentType,
		})
	}
	page.Data.Response.Categories = pageCategories

	//RESPONSE-ITEMS STRUCT MAPPING
	itemPage := []search.ContentItem{}
	for i, itemC := range respC.Items {
		descriptionC := itemC.Description
		itemPage = append(itemPage, search.ContentItem{
			Description: search.Description{
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
			},

			Type: itemC.Type,
			URI:  itemC.URI,
		})

		if descriptionC.Contact != nil {
			itemPage[i].Description.Contact = &search.Contact{
				Name:      descriptionC.Contact.Name,
				Telephone: descriptionC.Contact.Telephone,
				Email:     descriptionC.Contact.Email,
			}
		}

		if itemC.Matches != nil {
			matchesDescC := itemC.Matches.Description
			itemPage[i].Matches = &search.Matches{
				Description: search.MatchDescription{},
			}

			if matchesDescC.Summary != nil {
				matchesSummaryPage := []search.MatchDetails{}
				for _, summaryC := range *matchesDescC.Summary {
					matchesSummaryPage = append(matchesSummaryPage, search.MatchDetails{
						Value: summaryC.Value,
						Start: summaryC.Start,
						End:   summaryC.End,
					})
				}
				itemPage[i].Matches.Description.Summary = &matchesSummaryPage
			}

			if matchesDescC.Title != nil {
				matchesTitlePage := []search.MatchDetails{}
				for _, titleC := range *matchesDescC.Title {
					matchesTitlePage = append(matchesTitlePage, search.MatchDetails{
						Value: titleC.Value,
						Start: titleC.Start,
						End:   titleC.End,
					})
				}
				itemPage[i].Matches.Description.Title = &matchesTitlePage
			}

			if matchesDescC.Edition != nil {
				matchesEditionPage := []search.MatchDetails{}
				for _, editionC := range *matchesDescC.Edition {
					matchesEditionPage = append(matchesEditionPage, search.MatchDetails{
						Value: editionC.Value,
						Start: editionC.Start,
						End:   editionC.End,
					})
				}
				itemPage[i].Matches.Description.Edition = &matchesEditionPage
			}

			if matchesDescC.MetaDescription != nil {
				matchesMetaDescPage := []search.MatchDetails{}
				for _, metaDescC := range *matchesDescC.MetaDescription {
					matchesMetaDescPage = append(matchesMetaDescPage, search.MatchDetails{
						Value: metaDescC.Value,
						Start: metaDescC.Start,
						End:   metaDescC.End,
					})
				}
				itemPage[i].Matches.Description.MetaDescription = &matchesMetaDescPage
			}

			if matchesDescC.Keywords != nil {
				matchesKeywordsPage := []search.MatchDetails{}
				for _, keywordC := range *matchesDescC.Keywords {
					matchesKeywordsPage = append(matchesKeywordsPage, search.MatchDetails{
						Value: keywordC.Value,
						Start: keywordC.Start,
						End:   keywordC.End,
					})
				}
				itemPage[i].Matches.Description.Keywords = &matchesKeywordsPage
			}

			if matchesDescC.DatasetID != nil {
				matchesDatasetIDPage := []search.MatchDetails{}
				for _, datasetIDClient := range *matchesDescC.DatasetID {
					matchesDatasetIDPage = append(matchesDatasetIDPage, search.MatchDetails{
						Value: datasetIDClient.Value,
						Start: datasetIDClient.Start,
						End:   datasetIDClient.End,
					})
				}
				itemPage[i].Matches.Description.DatasetID = &matchesDatasetIDPage
			}
		}

		page.Data.Response.Suggestions = respC.Suggestions
	}
	page.Data.Response.Items = itemPage

	return page
}

func getFilterSortKeyList(query url.Values, categories []data.Category) []string {
	filterLocaliseKeyList := []string{}
	queryFilters := query["filter"]
	for _, filter := range queryFilters {
		for _, category := range categories {
			for _, contentType := range category.ContentTypes {
				if filter == contentType.Type {
					filterLocaliseKeyList = append(filterLocaliseKeyList, contentType.LocaliseKeyName)
				}
			}
		}
	}
	return filterLocaliseKeyList
}

func getSortLocaliseKey(query url.Values) (sortKey string) {
	querySort := query.Get("sort")
	for _, sort := range data.SortOptions {
		if querySort == sort.Query {
			sortKey = sort.LocaliseKeyName
		}
	}
	return sortKey
}

func getPagesToDisplay(currentPage int, totalPages int, url *url.URL) []model.PageToDisplay {
	var pagesToDisplay = make([]model.PageToDisplay, 0)
	startPage := currentPage - 2
	if currentPage <= 2 {
		startPage = 1
	} else {
		if (currentPage == totalPages-1) || (currentPage == totalPages) {
			startPage = totalPages - 4
		}
	}
	q := url.Query()
	query := q.Get("q")
	q.Del("q")
	q.Del("page")
	url.RawQuery = q.Encode()
	if url.RawQuery != "" {
		url.RawQuery = "&" + url.RawQuery
	}
	endPage := startPage + 4
	if totalPages < 5 {
		endPage = totalPages
	}
	for i := startPage; i <= endPage; i++ {
		pagesToDisplay = append(pagesToDisplay, model.PageToDisplay{
			PageNumber: i,
			URL:        "/search?q=" + query + url.RawQuery + "&page=" + strconv.Itoa(i),
		})
	}
	return pagesToDisplay
}
