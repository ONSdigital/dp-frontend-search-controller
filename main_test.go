package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	golog "log"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/features/steps"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(ctx *godog.ScenarioContext) {
	controllerComponent, err := steps.NewSearchControllerComponent()
	if err != nil {
		fmt.Printf("failed to create search controller component - error: %v", err)
		os.Exit(1)
	}
	apiFeature := controllerComponent.InitAPIFeature()

	url := fmt.Sprintf("http://%s%s", controllerComponent.Config.SiteDomain, controllerComponent.Config.BindAddr)
	uiFeature := componenttest.NewUIFeature(url)
	uiFeature.RegisterSteps(ctx)

	apiFeature.RegisterSteps(ctx)
	controllerComponent.RegisterSteps(ctx)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		uiFeature.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		uiFeature.Close()
		controllerComponent.Close()
		return ctx, nil
	})
}

func TestMainFunc(t *testing.T) {
	if *componentFlag {
		log.SetDestination(io.Discard, io.Discard)
		golog.SetOutput(io.Discard)
		defer func() {
			log.SetDestination(os.Stdout, os.Stderr)
			golog.SetOutput(os.Stdout)
		}()

		status := 0

		opts := godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
			Tags:   "@topicCache",
		}

		status = godog.TestSuite{
			Name:                "component_tests",
			ScenarioInitializer: InitializeScenario,
			Options:             &opts,
		}.Run()

		fmt.Println("=================================")
		fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")

		if status != 0 {
			t.FailNow()
		}

	} else {
		t.Skip("component flag required to run component tests")
	}
}
