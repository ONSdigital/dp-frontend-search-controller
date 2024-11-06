package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-search-controller/config"

	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
)

// read data aggregation
func (sh *SearchHandler) DataAggregation(cfg *config.Config, template string) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readDataAggregation(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, template)
	})
}
