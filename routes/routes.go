package routes

import (
	"context"
	"net/http"

	search "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"
	rend "github.com/ONSdigital/dp-renderer"
	zebedee "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// Clients - struct containing all the clients for the controller
type Clients struct {
	HealthCheckHandler func(w http.ResponseWriter, req *http.Request)
	Renderer           *rend.Render
	Search             *search.Client
	Zebedee            *zebedee.Client
}

// Setup registers routes for the service
func Setup(ctx context.Context, r *mux.Router, cfg *config.Config, zc handlers.ZebedeeClient, c Clients, rend handlers.RenderClient) {
	log.Info(ctx, "adding routes")
	r.StrictSlash(true).Path("/health").HandlerFunc(c.HealthCheckHandler)
	r.StrictSlash(true).Path("/search").Methods("GET").HandlerFunc(handlers.Read(cfg, zc, rend, c.Search))
}
