package service

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	render "github.com/ONSdigital/dp-renderer/v2"
	"github.com/ONSdigital/dp-renderer/v2/middleware/renderror"
	"github.com/justinas/alice"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-frontend-search-controller/assets"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	cachePrivate "github.com/ONSdigital/dp-frontend-search-controller/cache/private"
	cachePublic "github.com/ONSdigital/dp-frontend-search-controller/cache/public"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/routes"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	topic "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

// Service contains the healthcheck, server and serviceList for the frontend search controller
type Service struct {
	Cache              cache.List
	Config             *config.Config
	HealthCheck        HealthChecker
	routerHealthClient *health.Client
	Server             HTTPServer
	ServiceList        *ExternalServiceList
}

// New creates a new service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware
func (svc *Service) Init(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList) (err error) {
	log.Info(ctx, "initialising service")

	svc.Config = cfg
	svc.ServiceList = serviceList

	// Get health client for api router
	svc.routerHealthClient = serviceList.GetHealthClient("api-router", svc.Config.APIRouterURL)

	// Initialise clients
	clients := routes.Clients{
		Renderer: render.NewWithDefaultClient(assets.Asset, assets.AssetNames, svc.Config.PatternLibraryAssetsPath, svc.Config.SiteDomain),
		Search:   searchSDK.NewWithHealthClient(svc.routerHealthClient),
		Topic:    topic.NewWithHealthClient(svc.routerHealthClient),
		Zebedee:  zebedee.NewWithHealthClient(svc.routerHealthClient),
	}

	// Get healthcheck with checkers
	svc.HealthCheck, err = serviceList.GetHealthCheck(svc.Config, BuildTime, GitCommit, Version)
	if err != nil {
		log.Fatal(ctx, "failed to create health check", err)
		return err
	}
	if err = svc.registerCheckers(ctx, clients); err != nil {
		log.Error(ctx, "failed to register checkers", err)
		return err
	}
	clients.HealthCheckHandler = svc.HealthCheck.Handler

	// Initialise caching
	cache.CensusTopicID = svc.Config.CensusTopicID
	svc.Cache.CensusTopic, err = cache.NewTopicCache(ctx, &svc.Config.CacheCensusTopicUpdateInterval)
	if err != nil {
		log.Error(ctx, "failed to create census cache", err)
		return err
	}
	svc.Cache.DataTopic, err = cache.NewTopicCache(ctx, &svc.Config.CacheDataTopicUpdateInterval)
	if err != nil {
		log.Error(ctx, "failed to create data topics cache", err)
		return err
	}
	svc.Cache.Navigation, err = cache.NewNavigationCache(ctx, &svc.Config.CacheNavigationUpdateInterval)
	if err != nil {
		log.Error(ctx, "failed to create navigation cache", err, log.Data{"update_interval": svc.Config.CacheNavigationUpdateInterval})
		return err
	}

	if svc.Config.IsPublishing {
		svc.Cache.CensusTopic.AddUpdateFunc(cache.CensusTopicID, cachePrivate.UpdateCensusTopic(ctx, svc.Config.ServiceAuthToken, clients.Topic))
		if svc.Config.EnableTopicAggregationPages {
			svc.Cache.DataTopic.AddUpdateFuncs(cachePrivate.UpdateDataTopics(ctx, svc.Config.ServiceAuthToken, clients.Topic))
		}
	} else {
		svc.Cache.CensusTopic.AddUpdateFunc(cache.CensusTopicID, cachePublic.UpdateCensusTopic(ctx, clients.Topic))
		if svc.Config.EnableTopicAggregationPages {
			svc.Cache.DataTopic.AddUpdateFuncs(cachePublic.UpdateDataTopics(ctx, clients.Topic))
		}
	}

	for _, lang := range svc.Config.SupportedLanguages {
		navigationlangKey := svc.Cache.Navigation.GetCachingKeyForNavigationLanguage(lang)
		svc.Cache.Navigation.AddUpdateFunc(navigationlangKey, cachePublic.UpdateNavigationData(ctx, svc.Config, lang, clients.Topic))
	}

	// Initialise router
	r := mux.NewRouter()
	if svc.Config.OtelEnabled {
		r.Use(otelmux.Middleware(svc.Config.OTServiceName))
	}
	middleware := []alice.Constructor{
		renderror.Handler(clients.Renderer),
	}

	if svc.Config.OtelEnabled {
		middleware = append(middleware, otelhttp.NewMiddleware(svc.Config.OTServiceName))
	}

	newAlice := alice.New(middleware...).Then(r)
	routes.Setup(ctx, r, svc.Config, clients, svc.Cache)
	svc.Server = serviceList.GetHTTPServer(svc.Config.BindAddr, newAlice)

	return nil
}

// Run starts an initialised service
func (svc *Service) Run(ctx context.Context, svcErrors chan error) {
	log.Info(ctx, "Starting service", log.Data{"config": svc.Config})

	// Start healthcheck
	svc.HealthCheck.Start(ctx)

	// Start caching
	go svc.Cache.CensusTopic.StartUpdates(ctx, svcErrors)
	if svc.Config.EnableTopicAggregationPages {
		go svc.Cache.DataTopic.StartUpdates(ctx, svcErrors)
	}
	go svc.Cache.Navigation.StartUpdates(ctx, svcErrors)

	// Start HTTP server
	log.Info(ctx, "Starting server")
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			log.Fatal(ctx, "failed to start http listen and serve", err)
			svcErrors <- err
		}
	}()
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	log.Info(ctx, "commencing graceful shutdown")
	ctx, cancel := context.WithTimeout(ctx, svc.Config.GracefulShutdownTimeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		log.Info(ctx, "stop health checkers")
		svc.HealthCheck.Stop()

		// stop caching
		svc.Cache.CensusTopic.Close()
		if svc.Config.EnableTopicAggregationPages {
			svc.Cache.DataTopic.Close()
		}
		svc.Cache.Navigation.Close()

		// stop any incoming requests
		if err := svc.Server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context, c routes.Clients) (err error) {
	hasErrors := false

	if err = svc.HealthCheck.AddCheck("API router", svc.routerHealthClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add API router health checker", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}
