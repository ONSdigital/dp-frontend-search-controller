package routes

import (
	"context"
	"net/http"

	zebedee "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"
	rend "github.com/ONSdigital/dp-renderer/v2"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	topic "github.com/ONSdigital/dp-topic-api/sdk"

	"github.com/gorilla/mux"
)

// Clients - struct containing all the clients for the controller
type Clients struct {
	HealthCheckHandler func(w http.ResponseWriter, req *http.Request)
	Renderer           *rend.Render
	Search             *searchSDK.Client
	Topic              *topic.Client
	Zebedee            *zebedee.Client
}

// Setup registers routes for the service
func Setup(ctx context.Context, r *mux.Router, cfg *config.Config, c Clients, cacheList cache.List) {
	r.StrictSlash(true).Path("/health").HandlerFunc(c.HealthCheckHandler)
	r.StrictSlash(true).Path("/search").Methods("GET").HandlerFunc(handlers.Read(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "search"))

	r.StrictSlash(true).Path("/alladhocs").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "all-adhocs"))
	r.StrictSlash(true).Path("/datalist").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "home-datalist"))
	r.StrictSlash(true).Path("/publishedrequests").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "published-requests"))
	r.StrictSlash(true).Path("/staticlist").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "home-list"))
	r.StrictSlash(true).Path("/topicspecificmethodology").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "home-methodology"))
	r.StrictSlash(true).Path("/timeseriestool").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "time-series-tool"))
	r.StrictSlash(true).Path("/publications").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "home-publications"))
	r.StrictSlash(true).Path("/allmethodologies").Methods("GET").HandlerFunc(handlers.ReadDataAggregation(cfg, c.Zebedee, c.Renderer, c.Search, cacheList, "all-methodologies"))

	r.StrictSlash(true).Path("/census/find-a-dataset").Methods("GET").HandlerFunc(handlers.ReadFindDataset(cfg, c.Zebedee, c.Renderer, c.Search, cacheList))
}
