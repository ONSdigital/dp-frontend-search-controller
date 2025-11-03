package routes

import (
	"context"
	"net/http"

	rend "github.com/ONSdigital/dis-design-system-go"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"
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
	sh := handlers.NewSearchHandler(c.Renderer, c.Search, c.Topic, c.Zebedee, cfg, cacheList)

	r.StrictSlash(true).Path("/health").HandlerFunc(c.HealthCheckHandler)
	r.StrictSlash(true).Path("/search").Methods("GET").HandlerFunc(sh.Search(cfg))

	if sh.EnableAggregationPages {
		r.StrictSlash(true).Path("/alladhocs").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "all-adhocs"))
		r.StrictSlash(true).Path("/datalist").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "home-datalist"))
		r.StrictSlash(true).Path("/publishedrequests").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "published-requests"))
		r.StrictSlash(true).Path("/staticlist").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "home-list"))
		r.StrictSlash(true).Path("/topicspecificmethodology").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "home-methodology"))
		r.StrictSlash(true).Path("/timeseriestool").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "time-series-tool"))
		r.StrictSlash(true).Path("/publications").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "home-publications"))
		r.StrictSlash(true).Path("/allmethodologies").Methods("GET").HandlerFunc(sh.DataAggregation(cfg, "all-methodologies"))

		if sh.EnableTopicAggregationPages {
			// handle dynamic aggregated data topic pages
			r.StrictSlash(true).Path("/{topicsPath:.*}/datalist").Methods("GET").HandlerFunc(sh.DataAggregationWithTopics(cfg, "home-datalist"))
			r.StrictSlash(true).Path("/{topicsPath:.*}/publications").Methods("GET").HandlerFunc(sh.DataAggregationWithTopics(cfg, "home-publications"))
			r.StrictSlash(true).Path("/{topicsPath:.*}/staticlist").Methods("GET").HandlerFunc(sh.DataAggregationWithTopics(cfg, "home-list"))
			r.StrictSlash(true).Path("/{topicsPath:.*}/topicspecificmethodology").Methods("GET").HandlerFunc(sh.DataAggregationWithTopics(cfg, "home-methodology"))
		}
	}

	r.StrictSlash(true).Path("/{uri:.*}/previousreleases").Methods("GET").HandlerFunc(sh.PreviousReleases(cfg))
	r.StrictSlash(true).Path("/{uri:.*}/relateddata").Methods("GET").HandlerFunc(sh.RelatedData(cfg))

	r.StrictSlash(true).Path("/census/find-a-dataset").Methods("GET").HandlerFunc(sh.FindDataset(cfg))
}
