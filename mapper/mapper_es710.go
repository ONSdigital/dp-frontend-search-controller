package mapper

import (
	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	model "github.com/ONSdigital/dp-frontend-search-controller/model"
)

func mapResponse(page *model.SearchPage, respC searchC.Response, categories []data.Category) {
	page.Data.Response.Count = respC.Count

	mapResponseCategories(page, categories)

	mapResponseItems(page, respC)

	page.Data.Response.Suggestions = respC.Suggestions
	page.Data.Response.AdditionalSuggestions = respC.AdditionalSuggestions
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

	item.Description = model.Description{
		DatasetID:       itemC.DatasetID,
		Language:        itemC.Language,
		MetaDescription: itemC.MetaDescription,
		ReleaseDate:     itemC.ReleaseDate,
		Summary:         itemC.Summary,
		Title:           itemC.Title,
	}

	if len(itemC.Keywords) != 0 {
		item.Description.Keywords = &itemC.Keywords
	} else {
		item.Description.Keywords = nil
	}
}

func mapItemHighlight(item *model.ContentItem, itemC searchC.ContentItem) {
	highlightC := itemC.Highlight

	if highlightC != nil {
		item.Description.Highlight = model.Highlight{
			DatasetID:       highlightC.DatasetID,
			Edition:         highlightC.Edition,
			Keywords:        highlightC.Keywords,
			MetaDescription: highlightC.MetaDescription,
			Summary:         highlightC.Summary,
			Title:           highlightC.Title,
		}
	} else {
		item.Description.Highlight = model.Highlight{}
	}
}
