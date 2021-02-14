package handlers

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/log.go/log"
)

func mapSubFilterTypes(ctx context.Context, query url.Values) (apiQuery url.Values, err error) {
	apiQuery = updateQueryWithOffset(ctx, query)
	apiQuery, err = url.ParseQuery(apiQuery.Encode())
	if err != nil {
		log.Event(ctx, "failed to parse copy of query for mapping filter types", log.Error(err), log.ERROR)
		return nil, err
	}
	filters := apiQuery["filter"]
	if len(filters) > 0 {
		var newFilters = make([]string, 0)
		for _, fType := range filters {
			found := false
		categoryLoop:
			for _, category := range data.Categories {
				for _, contentType := range category.ContentTypes {
					if fType == contentType.Type {
						found = true
						newFilters = append(newFilters, contentType.SubTypes...)
						break categoryLoop
					}
				}
			}
			if !found {
				return nil, errFilterType
			}
		}
		apiQuery.Del("filter")
		apiQuery.Set("content_type", strings.Join(newFilters, ","))
	}
	return apiQuery, nil
}

func getCategoriesTypesCount(ctx context.Context, apiQuery url.Values, searchC SearchClient) (categories []data.Category, err error) {
	//Remove filter to get count of all types for the query from the client
	apiQuery.Del("content_type")
	countResp, err := searchC.GetSearch(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "getting search query count from client failed", log.Error(err), log.ERROR)
		return nil, err
	}
	categories = data.GetAllCategories()
	for _, responseType := range countResp.ContentTypes {
		foundFilter := false
	categoryLoop:
		for i, category := range categories {
			for j, contentType := range category.ContentTypes {
				for _, subType := range contentType.SubTypes {
					if responseType.Type == subType {
						categories[i].Count += responseType.Count
						categories[i].ContentTypes[j].Count += responseType.Count
						foundFilter = true
						break categoryLoop
					}
				}
			}
		}
		if !foundFilter {
			return nil, errors.New("filter type from client not available in data.go")
		}
	}
	return categories, nil
}
