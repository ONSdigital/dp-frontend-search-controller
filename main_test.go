package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/features/steps"
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

	apiFeature := componenttest.NewAPIFeature(controllerComponent.InitialiseService)

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		controllerComponent.Reset()
	})

	apiFeature.RegisterSteps(ctx)
	controllerComponent.RegisterSteps(ctx)

	ctx.AfterScenario(func(*godog.Scenario, error) {
		controllerComponent.Close()
	})
}

func TestMain(t *testing.T) {
	if *componentFlag {
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

		os.Exit(status)

	} else {
		t.Skip("component flag required to run component tests")
	}
}
