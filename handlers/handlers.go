package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-frontend-search-controller/mapper"

	search "github.com/ONSdigital/dp-api-clients-go/site-search"
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
	fmt.Println(templateHTML)
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
	apiQuery := mapFilterTypes(query)
	resp, err := searchC.GetSearch(ctx, apiQuery)
	if err != nil {
		log.Event(ctx, "getting search response from client failed", log.Error(err), log.ERROR)
		setStatusCode(req, w, err)
		return
	}
	err = getSearchPage(w, req, rendC, query, resp)
	if err != nil {
		log.Event(ctx, "getting search page failed", log.Error(err), log.ERROR)
	}
	return
}

func mapFilterTypes(query url.Values) url.Values {
	//Filter Mapping
	//If filters are added or removed in the map, make sure to do the same in the defaultContentTypes variable in dp-setup-query
	filterTypes := map[string][]string{
		"bulletin":              {"bulletin"},
		"article":               {"article", "article_download", "static_article"},
		"compendia":             {"compendium_landing_page", "compendium_chapter"},
		"time_series":           {"timeseries"},
		"datasets":              {"dataset", "dataset_landing_page", "compendium_data", "reference_tables", "timeseries_dataset"},
		"user_requested_data":   {"static_adhoc"},
		"methodology":           {"static_methodology", "static_methodology_download", "static_qmi"},
		"corporate_information": {"static_foi", "static_page", "static_landing_page", "static_article"},
	}

	filters := query["filter"]
	if len(filters) > 0 {
		var newFilters []string
		for _, fType := range filters {
			newFilters = append(newFilters, filterTypes[fType]...)
		}
		query["filter"] = newFilters
	}

	return query
}
