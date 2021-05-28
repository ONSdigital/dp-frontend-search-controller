package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/log.go/log"
)

const (
	serviceName = "dp-frontend-search-controller"
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	// Create error channel for os signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create service initialiser and an error channel for service errors
	svcList := service.NewServiceList(&service.Init{})
	svcErrors := make(chan error, 1)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.ERROR, log.Error(err))
		return err
	}
	log.Event(ctx, "got service configuration", log.INFO, log.Data{"config": cfg})

	// Run service
	srv := service.New()
	if err := srv.Init(ctx, cfg, svcList); err != nil {
		log.Event(ctx, "failed to initialise service", log.ERROR, log.Error(err))
		return err
	}
	srv.Start(ctx, svcErrors)

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Event(ctx, "service error received", log.ERROR, log.Error(err))
	case sig := <-signals:
		log.Event(ctx, "os signal received", log.Data{"signal": sig}, log.INFO)
	}

	return srv.Close(ctx)
}
