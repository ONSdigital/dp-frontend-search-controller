package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
)

func (sh *SearchHandler) FindDataset(cfg *config.Config) http.HandlerFunc {
	return dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		readFindDataset(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList)
	})
}
