package routes

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/renderer"
	search "github.com/ONSdigital/dp-api-clients-go/site-search"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//Clients - struct containing all the clients for the controller
type Clients struct {
	HealthCheckHandler func(w http.ResponseWriter, req *http.Request)
	Renderer           *renderer.Renderer
	Search             *search.Client
}

// Setup registers routes for the service
func Setup(ctx context.Context, r *mux.Router, cfg *config.Config, c Clients) {
	log.Event(ctx, "adding routes")

	r.StrictSlash(true).Path("/health").HandlerFunc(c.HealthCheckHandler)
	r.StrictSlash(true).Path("/search").Methods("GET").HandlerFunc(handlers.Read(cfg, c.Renderer, c.Search))
}
