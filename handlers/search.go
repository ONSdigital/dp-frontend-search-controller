package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	dphandlers "github.com/ONSdigital/dp-net/v2/handlers"
)

// Read Handler
func (sh *SearchHandler) Search(cfg *config.Config, template string) http.HandlerFunc {
	oldHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, template, false)
	})

	newHandler := dphandlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, accessToken string) {
		read(w, req, cfg, sh.ZebedeeClient, sh.Renderer, sh.SearchClient, accessToken, collectionID, lang, sh.CacheList, template, true)
	})

	return cookies.Handler(cfg.ABTest.Enabled, newHandler, oldHandler, cfg.ABTest.Percentage, cfg.ABTest.AspectID, cfg.SiteDomain, cfg.ABTest.Exit)
}
