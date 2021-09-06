package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	searchCli "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/data"
	"github.com/ONSdigital/dp-frontend-search-controller/mapper"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/v2/log"
)

// Read Handler
func Read(cfg *config.Config, rendC RenderClient, searchC SearchClient) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, rendC, searchC, accessToken, collectionID, lang)
	})
}

func read(w http.ResponseWriter, req *http.Request, cfg *config.Config, rendC RenderClient, searchC SearchClient, accessToken, collectionID, lang string) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	urlQuery := req.URL.Query()

	validatedQueryParams, err := data.ReviewQuery(ctx, cfg, urlQuery)
	if err != nil {
		log.Error(ctx, "unable to review query", err)
		setStatusCode(w, req, err)
		return
	}

	apiQuery := data.GetSearchAPIQuery(validatedQueryParams)

	var searchResp searchCli.Response
	var respErr error
	var departmentResp searchCli.Department
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		searchResp, respErr = searchC.GetSearch(ctx, accessToken, "", collectionID, apiQuery)
		if respErr != nil {
			logData := log.Data{"api query passed to search-api": apiQuery}
			log.Error(ctx, "getting search response from client failed", respErr, logData)
			cancel()
			return
		}
	}()
	go func() {
		defer wg.Done()
		var deptErr error
		departmentResp, deptErr = searchC.GetDepartments(ctx, accessToken, "", collectionID, apiQuery)
		if deptErr != nil {
			logData := log.Data{"api query passed to search-api": apiQuery}
			log.Error(ctx, "getting deartment response from client failed", deptErr, logData)
			return
		}
	}()

	wg.Wait()
	if respErr != nil {
		setStatusCode(w, req, respErr)
		return
	}

	// TO-DO: Until API handles aggregration on datatypes (e.g. bulletins, article), we need to make a second request

	err = validateCurrentPage(ctx, cfg, validatedQueryParams, searchResp.Count)
	if err != nil {
		log.Error(ctx, "unable to validate current page", err)
		setStatusCode(w, req, err)
		return
	}

	categories, err := getCategoriesTypesCount(ctx, accessToken, collectionID, apiQuery, searchC)
	if err != nil {
		log.Error(ctx, "getting categories, types and its counts failed", err)
		setStatusCode(w, req, err)
		return
	}

	err = getSearchPage(w, req, cfg, rendC, validatedQueryParams, categories, searchResp, departmentResp, lang)
	if err != nil {
		log.Error(ctx, "getting search page failed", err)
		setStatusCode(w, req, err)
		return
	}
}

// validateCurrentPage checks if the current page exceeds the total pages which is a bad request
func validateCurrentPage(ctx context.Context, cfg *config.Config, validatedQueryParams data.SearchURLParams, resultsCount int) error {

	if resultsCount > 0 {
		totalPages := data.GetTotalPages(cfg, validatedQueryParams.Limit, resultsCount)

		if validatedQueryParams.CurrentPage > totalPages {
			err := errs.ErrPageExceedsTotalPages
			log.Error(ctx, "current page exceeds total pages", err)

			return err
		}
	}

	return nil
}

// getCategoriesTypesCount removes the filters and communicates with the search api again to retrieve the number of search results for each filter categories and subtypes
func getCategoriesTypesCount(ctx context.Context, accessToken, collectionID string, apiQuery url.Values, searchC SearchClient) ([]data.Category, error) {
	//Remove filter to get count of all types for the query from the client
	apiQuery.Del("content_type")

	countResp, err := searchC.GetSearch(ctx, accessToken, "", collectionID, apiQuery)
	if err != nil {
		logData := log.Data{"api query passed to search-api": apiQuery}
		log.Error(ctx, "getting search query count from client failed", err, logData)
		return nil, err
	}

	categories := data.GetCategories()

	setCountToCategories(ctx, countResp, categories)

	return categories, nil
}

func setCountToCategories(ctx context.Context, countResp searchCli.Response, categories []data.Category) {
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
			log.Warn(ctx, "unrecognised filter type returned from api")
		}
	}
}

// getSearchPage talks to the renderer to get the search page
func getSearchPage(w http.ResponseWriter, req *http.Request, cfg *config.Config, rendC RenderClient, validatedQueryParams data.SearchURLParams, categories []data.Category, resp searchCli.Response, departments searchCli.Department, lang string) error {
	ctx := req.Context()

	m := mapper.CreateSearchPage(cfg, validatedQueryParams, categories, resp, departments, lang)

	b, err := json.Marshal(m)
	if err != nil {
		logData := log.Data{"search response": m}
		log.Error(ctx, "unable to marshal search response", err, logData)
		setStatusCode(w, req, err)
		return err
	}

	templateHTML, err := rendC.Do("search", b)
	if err != nil {
		logData := log.Data{"retrieving search template for response": b}
		log.Error(ctx, "getting template from renderer search failed", err, logData)
		setStatusCode(w, req, err)
		return err
	}

	if _, err := w.Write(templateHTML); err != nil {
		logData := log.Data{"search template": templateHTML}
		log.Error(ctx, "error on write of search template", err, logData)
		setStatusCode(w, req, err)
		return err
	}

	return err
}

func setStatusCode(w http.ResponseWriter, req *http.Request, err error) {
	status := http.StatusInternalServerError

	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}

	if errs.BadRequestMap[err] {
		status = http.StatusBadRequest
	}

	log.Error(req.Context(), "setting-response-status", err)

	w.WriteHeader(status)
}
