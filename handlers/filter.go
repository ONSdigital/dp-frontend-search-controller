package handlers

import (
	"context"
	"errors"
	"net/url"
	"strings"

	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/log.go/log"
)

func mapFilterTypes(ctx context.Context, query url.Values) (apiQuery url.Values, err error) {
	apiQuery, err = url.ParseQuery(query.Encode())
	if err != nil {
		log.Event(ctx, "failed to parse copy of query for mapping filter types", log.Error(err), log.ERROR)
		return nil, err
	}
	filters := apiQuery["filter"]
	if len(filters) > 0 {
		var newFilters []string
		for _, fType := range filters {
			found := false
		categoryLoop:
			for _, category := range data.Category {
				for _, searchType := range category {
					if fType == searchType.QueryType {
						found = true
						newFilters = append(newFilters, searchType.SubTypes...)
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

func mapCountFilterTypes(ctx context.Context, apiQuery url.Values, searchC SearchClient) (mappedContentType []search.ContentType, err error) {
	//Remove filter to get count of all types for the query from the client
	apiQuery.Del("content_type")
	countResp, err := searchC.GetSearch(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "getting search query count from client failed", log.Error(err), log.ERROR)
		return nil, err
	}
	countFilter := make(map[string]int)
	for _, contentType := range countResp.ContentTypes {
		foundFilter := false
		for _, category := range data.Category {
			for _, searchType := range category {
				for _, filter := range searchType.SubTypes {
					if filter == contentType.Type {
						countFilter[searchType.QueryType] += contentType.Count
						foundFilter = true
					}
				}
			}
		}
		if !foundFilter {
			return nil, errors.New("filter type from client not available in filterTypes map")
		}
	}
	for k, v := range countFilter {
		mappedContentType = append(mappedContentType, search.ContentType{
			Type:  k,
			Count: v,
		})
	}

	return mappedContentType, nil
}
