package mapper

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	searchC "github.com/ONSdigital/dp-api-clients-go/site-search"
	model "github.com/ONSdigital/dp-frontend-models/model/search"
	"github.com/ONSdigital/log.go/log"
)

// CreateSearchPage maps type cookies.Policy to model.Page
func CreateSearchPage(ctx context.Context, query url.Values, respC searchC.Response) (page model.Page) {
	var err error

	// SEARCH STRUCT MAPPING
	page.Data.Query = query.Get("q")

	contentStr := query.Get("content_type")
	contentStr = strings.TrimLeft(contentStr, "[")
	contentStr = strings.TrimRight(contentStr, "]")
	page.Data.Filter = strings.Split(contentStr, ",")

	page.Data.Sort = query.Get("sort_order")

	if query.Get("limit") != "" {
		page.Data.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			log.Event(ctx, "unable to convert search limit to int", log.Error(err), log.ERROR)
		}
	}

	if query.Get("offset") != "" {
		page.Data.Offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			log.Event(ctx, "unable to convert search offset to int", log.Error(err), log.ERROR)
		}
	}

	//RESPONSE STRUCT MAPPING
	page.Data.Response.Count = respC.Count

	pageContentType := page.Data.Response.ContentTypes
	for i, contentTypeC := range respC.ContentTypes {
		pageContentType[i].Type = contentTypeC.Type
		pageContentType[i].Count = contentTypeC.Count
	}

	//RESPONSE-ITEMS STRUCT MAPPING
	for i, itemC := range respC.Items {
		itemPage := page.Data.Response.Items[i]

		descriptionPage := itemPage.Description
		descriptionC := itemC.Description

		descriptionPage.Contact.Name = descriptionC.Contact.Name
		descriptionPage.Contact.Telephone = descriptionC.Contact.Telephone
		descriptionPage.Contact.Email = descriptionC.Contact.Email
		descriptionPage.DatasetID = descriptionC.DatasetID
		descriptionPage.Edition = descriptionC.Edition
		descriptionPage.Headline1 = descriptionC.Headline1
		descriptionPage.Headline2 = descriptionC.Headline2
		descriptionPage.Headline3 = descriptionC.Headline3
		descriptionPage.Keywords = descriptionC.Keywords
		descriptionPage.LatestRelease = descriptionC.LatestRelease
		descriptionPage.Language = descriptionC.Language
		descriptionPage.MetaDescription = descriptionC.MetaDescription
		descriptionPage.NationalStatistic = descriptionC.NationalStatistic
		descriptionPage.NextRelease = descriptionC.NextRelease
		descriptionPage.PreUnit = descriptionC.PreUnit
		descriptionPage.ReleaseDate = descriptionC.ReleaseDate
		descriptionPage.Source = descriptionC.Source
		descriptionPage.Summary = descriptionC.Summary
		descriptionPage.Title = descriptionC.Title
		descriptionPage.Unit = descriptionC.Unit

		itemPage.Type = itemC.Type
		itemPage.URI = itemC.URI

		matchesDescPage := itemPage.Matches.Description
		matchesDescC := itemC.Matches.Description

		matchesSummaryPage := *matchesDescPage.Summary
		for j, summaryC := range *matchesDescC.Summary {
			matchesSummaryPage[j].Value = summaryC.Value
			matchesSummaryPage[j].Start = summaryC.Start
			matchesSummaryPage[j].End = summaryC.End
		}

		matchesTitlePage := *matchesDescPage.Title
		for j, titleC := range *matchesDescC.Title {
			matchesTitlePage[j].Value = titleC.Value
			matchesTitlePage[j].Start = titleC.Start
			matchesTitlePage[j].End = titleC.End
		}

		matchesEditionPage := *matchesDescPage.Edition
		for j, editionC := range *matchesDescC.Edition {
			matchesEditionPage[j].Value = editionC.Value
			matchesEditionPage[j].Start = editionC.Start
			matchesEditionPage[j].End = editionC.End
		}

		matchesMetaDescPage := *matchesDescPage.MetaDescription
		for j, metaDescC := range *matchesDescC.MetaDescription {
			matchesMetaDescPage[j].Value = metaDescC.Value
			matchesMetaDescPage[j].Start = metaDescC.Start
			matchesMetaDescPage[j].End = metaDescC.End
		}

		matchesKeywordsPage := *matchesDescPage.Keywords
		for j, keywordC := range *matchesDescC.Keywords {
			matchesKeywordsPage[j].Value = keywordC.Value
			matchesKeywordsPage[j].Start = keywordC.Start
			matchesKeywordsPage[j].End = keywordC.End
		}

		matchesDatasetIDPage := *matchesDescPage.DatasetID
		for j, datasetIDClient := range *matchesDescC.DatasetID {
			matchesDatasetIDPage[j].Value = datasetIDClient.Value
			matchesDatasetIDPage[j].Start = datasetIDClient.Start
			matchesDatasetIDPage[j].End = datasetIDClient.End
		}
	}

	page.Data.Response.Suggestions = respC.Suggestions

	return page
}
