package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/log.go/log"
)

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Error() string
	Code() int
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	Do(string, []byte) ([]byte, error)
}

// SearchClient is an interface with methods required for a search client
type SearchClient interface {
	GetSearch(ctx context.Context, query url.Values) (r search.Response, err error)
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	if err.Error() == "invalid filter type given" {
		status = http.StatusBadRequest
	}
	if err.Error() == "current page exceeds total pages" {
		status = http.StatusBadRequest
	}
	log.Event(req.Context(), "setting-response-status", log.Error(err), log.ERROR)
	w.WriteHeader(status)
}

var marshal = func(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

var writeResponse = func(w http.ResponseWriter, templateHTML []byte) (int, error) {
	return w.Write(templateHTML)
}

// getSearchPage talks to the renderer to get the search page
func getSearchPage(w http.ResponseWriter, req *http.Request, rendC RenderClient, url *url.URL, resp search.Response, categories []data.Category, paginationQuery *data.PaginationQuery) error {
	ctx := req.Context()
	m := mapper.CreateSearchPage(ctx, url, resp, categories, paginationQuery)
	b, err := marshal(m)
	if err != nil {
		log.Event(ctx, "unable to marshal search response", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return err
	}
	templateHTML, err := rendC.Do("search", b)
	if err != nil {
		log.Event(ctx, "getting template from renderer search failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return err
	}
	if _, err := writeResponse(w, templateHTML); err != nil {
		log.Event(ctx, "error on write of search template", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return err
	}
	return err
}

func isCurrentPageLessThanTotalPages(ctx context.Context, paginationQuery *data.PaginationQuery, resp search.Response) (bool, error) {
	totalPages := (resp.Count + paginationQuery.Limit - 1) / paginationQuery.Limit
	if paginationQuery.CurrentPage > totalPages {
		return false, errors.New("current page exceeds total pages")
	}
	return true, nil
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

// Read Handler
func Read(rendC RenderClient, searchC SearchClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		read(w, req, rendC, searchC)
	}
}

func read(w http.ResponseWriter, req *http.Request, rendC RenderClient, searchC SearchClient) {
	ctx := req.Context()
	url, paginationQuery := data.SetDefaultQueries(ctx, req.URL)
	apiQuery, err := data.MapSubFilterTypes(ctx, url.Query())
	if err != nil {
		log.Event(ctx, "mapping sub filter types to query failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	resp, err := searchC.GetSearch(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "getting search response from client failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	validCurrentPage, err := isCurrentPageLessThanTotalPages(ctx, paginationQuery, resp)
	if !validCurrentPage {
		log.Event(ctx, "given page is not valid", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	categories, err := getCategoriesTypesCount(ctx, apiQuery, searchC)
	if err != nil {
		log.Event(ctx, "mapping count filter types failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	err = getSearchPage(w, req, rendC, url, resp, categories, paginationQuery)
	if err != nil {
		log.Event(ctx, "getting search page failed", log.Error(err), log.ERROR)
	}
	return
}
