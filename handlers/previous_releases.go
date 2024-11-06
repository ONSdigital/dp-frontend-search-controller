package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
)

// ReadPreviousReleases handles previous releases page
func (sh *SearchHandler) PreviousReleases(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readPreviousReleases(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList)
	})
}
