package main

import (
	"flag"
	"fmt"
	"io"
	golog "log"
	"os"
	"testing"

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

	apiFeature.RegisterSteps(ctx)
	controllerComponent.RegisterSteps(ctx)

	ctx.AfterScenario(func(*godog.Scenario, error) {
		controllerComponent.Close()
	})
}

func TestMain(t *testing.T) {
	if *componentFlag {
		log.SetDestination(io.Discard, io.Discard)
		golog.SetOutput(io.Discard)
		defer func() {
			log.SetDestination(os.Stdout, os.Stderr)
			golog.SetOutput(os.Stdout)
		}()

		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
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
