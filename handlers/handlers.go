package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	"github.com/ONSdigital/log.go/log"
)

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

// updateQueryWithOffset - removes page key and adds offset key to query to be then passed to dp-search-query
func updateQueryWithOffset(ctx context.Context, query url.Values) url.Values {
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		log.Event(ctx, "unable to convert search page to int - set to default 1", log.INFO)
		page = 1
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		log.Event(ctx, "unable to convert search limit to int - set to default 10", log.INFO)
		limit = 10
	}
	offset := strconv.Itoa(((page - 1) * limit))
	updateQuery := query
	updateQuery.Set("offset", offset)
	updateQuery.Del("page")
	return updateQuery
}

// getSearchPage talks to the renderer to get the search page
func getSearchPage(w http.ResponseWriter, req *http.Request, rendC RenderClient, url *url.URL, resp search.Response, categories []data.Category) error {
	ctx := req.Context()
	m := mapper.CreateSearchPage(ctx, url, resp, categories)
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
	url := req.URL
	apiQuery, err := mapSubFilterTypes(ctx, url.Query())
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
	categories, err := getCategoriesTypesCount(ctx, apiQuery, searchC)
	if err != nil {
		log.Event(ctx, "mapping count filter types failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	err = getSearchPage(w, req, rendC, url, resp, categories)
	if err != nil {
		log.Event(ctx, "getting search page failed", log.Error(err), log.ERROR)
	}
	return
}
