package routes

import (
	"context"
	"net/http"

	zebedee "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"
	rend "github.com/ONSdigital/dp-renderer"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	topic "github.com/ONSdigital/dp-topic-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"
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
	log.Info(ctx, "adding routes")
	r.StrictSlash(true).Path("/health").HandlerFunc(c.HealthCheckHandler)
	r.StrictSlash(true).Path("/search").Methods("GET").HandlerFunc(handlers.Read(cfg, c.Zebedee, c.Renderer, c.Search, cacheList))
}
