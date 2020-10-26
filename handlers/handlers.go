package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-frontend-search-controller/mapper"

	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/log.go/log"
)

//Filter Mapping
//If filters are added or removed in the map, make sure to do the same in the defaultContentTypes variable in dp-setup-query
var filterTypes = map[string]string{
	"bulletin":              "bulletin",
	"article":               "article,article_download",
	"compendia":             "compendium_landing_page,compendium_chapter",
	"time_series":           "timeseries",
	"datasets":              "dataset,dataset_landing_page,compendium_data,reference_tables,timeseries_dataset",
	"user_requested_data":   "static_adhoc",
	"methodology":           "static_methodology,static_methodology_download,static_qmi",
	"corporate_information": "static_foi,static_page,static_landing_page,static_article",
}

var errFilterType = errors.New("invalid filter type given")

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
func getSearchPage(w http.ResponseWriter, req *http.Request, rendC RenderClient, query url.Values, resp search.Response) error {
	ctx := req.Context()
	m := mapper.CreateSearchPage(ctx, query, resp)
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

// Read Handler
func Read(rendC RenderClient, searchC SearchClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		read(w, req, rendC, searchC)
	}
}

func read(w http.ResponseWriter, req *http.Request, rendC RenderClient, searchC SearchClient) {
	ctx := req.Context()
	query := req.URL.Query()
	apiQuery, err := mapFilterTypes(ctx, query)
	if err != nil {
		log.Event(ctx, "mapping filter types failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	resp, err := searchC.GetSearch(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "getting search response from client failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	resp.ContentTypes, err = mapCountFilterTypes(ctx, apiQuery, searchC)
	if err != nil {
		log.Event(ctx, "mapping count filter types failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	err = getSearchPage(w, req, rendC, query, resp)
	if err != nil {
		log.Event(ctx, "getting search page failed", log.Error(err), log.ERROR)
	}
	return
}

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
			if _, ok := filterTypes[fType]; !ok {
				return nil, errFilterType
			}
			newFilters = append(newFilters, filterTypes[fType])
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
		for k, v := range filterTypes {
			mapfilters := strings.Split(v, ",")
			for _, filter := range mapfilters {
				if filter == contentType.Type {
					countFilter[k] += contentType.Count
					foundFilter = true
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
