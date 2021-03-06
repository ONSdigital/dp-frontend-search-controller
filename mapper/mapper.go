package mapper

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	model "github.com/ONSdigital/dp-frontend-models/model/search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/log.go/log"
)

// CreateSearchPage maps type cookies.Policy to model.Page
func CreateSearchPage(ctx context.Context, query url.Values, respC searchC.Response, categories []data.Category) (page model.Page) {
	// SEARCH STRUCT MAPPING
	var err error
	page.Metadata.Title = "Search"
	page.SearchDisabled = true
	page.Data.Query = query.Get("q")

	page.Data.Filter.Query = query["filter"]
	page.Data.Filter.Options = []string{"Publication", "Data", "Other"}

	page.Data.Sort.Query = query.Get("sort")
	page.Data.Sort.LocaliseFilterKeys = getFilterSortKeyList(query, categories)
	page.Data.Sort.FilterText = strings.Replace(strings.ToLower(page.Data.Sort.Query), "-", " ", 1)

	if query.Get("limit") != "" {
		page.Data.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			log.Event(ctx, "unable to convert search limit to int - default to limit 10", log.INFO)
			page.Data.Limit = 10
		}
	}

	if query.Get("offset") != "" {
		page.Data.Offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			log.Event(ctx, "unable to convert search offset to int - default to offset 0", log.INFO)
			page.Data.Offset = 0
		}
	}

	//RESPONSE STRUCT MAPPING
	page.Data.Response.Count = respC.Count

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

	//RESPONSE-ITEMS STRUCT MAPPING
	itemPage := []model.ContentItem{}
	for i, itemC := range respC.Items {
		descriptionC := itemC.Description
		itemPage = append(itemPage, model.ContentItem{
			Description: model.Description{
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
			itemPage[i].Description.Contact = &model.Contact{
				Name:      descriptionC.Contact.Name,
				Telephone: descriptionC.Contact.Telephone,
				Email:     descriptionC.Contact.Email,
			}
		}

		if itemC.Matches != nil {
			matchesDescC := itemC.Matches.Description
			itemPage[i].Matches = &model.Matches{
				Description: model.MatchDescription{},
			}

			if matchesDescC.Summary != nil {
				matchesSummaryPage := []model.MatchDetails{}
				for _, summaryC := range *matchesDescC.Summary {
					matchesSummaryPage = append(matchesSummaryPage, model.MatchDetails{
						Value: summaryC.Value,
						Start: summaryC.Start,
						End:   summaryC.End,
					})
				}
				itemPage[i].Matches.Description.Summary = &matchesSummaryPage
			}

			if matchesDescC.Title != nil {
				matchesTitlePage := []model.MatchDetails{}
				for _, titleC := range *matchesDescC.Title {
					matchesTitlePage = append(matchesTitlePage, model.MatchDetails{
						Value: titleC.Value,
						Start: titleC.Start,
						End:   titleC.End,
					})
				}
				itemPage[i].Matches.Description.Title = &matchesTitlePage
			}

			if matchesDescC.Edition != nil {
				matchesEditionPage := []model.MatchDetails{}
				for _, editionC := range *matchesDescC.Edition {
					matchesEditionPage = append(matchesEditionPage, model.MatchDetails{
						Value: editionC.Value,
						Start: editionC.Start,
						End:   editionC.End,
					})
				}
				itemPage[i].Matches.Description.Edition = &matchesEditionPage
			}

			if matchesDescC.MetaDescription != nil {
				matchesMetaDescPage := []model.MatchDetails{}
				for _, metaDescC := range *matchesDescC.MetaDescription {
					matchesMetaDescPage = append(matchesMetaDescPage, model.MatchDetails{
						Value: metaDescC.Value,
						Start: metaDescC.Start,
						End:   metaDescC.End,
					})
				}
				itemPage[i].Matches.Description.MetaDescription = &matchesMetaDescPage
			}

			if matchesDescC.Keywords != nil {
				matchesKeywordsPage := []model.MatchDetails{}
				for _, keywordC := range *matchesDescC.Keywords {
					matchesKeywordsPage = append(matchesKeywordsPage, model.MatchDetails{
						Value: keywordC.Value,
						Start: keywordC.Start,
						End:   keywordC.End,
					})
				}
				itemPage[i].Matches.Description.Keywords = &matchesKeywordsPage
			}

			if matchesDescC.DatasetID != nil {
				matchesDatasetIDPage := []model.MatchDetails{}
				for _, datasetIDClient := range *matchesDescC.DatasetID {
					matchesDatasetIDPage = append(matchesDatasetIDPage, model.MatchDetails{
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
