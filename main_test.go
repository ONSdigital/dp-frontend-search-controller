package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

func InitializeScenario(ctx *godog.ScenarioContext) {
	fakeAPIRouter := steps.NewFakeAPI()

	controllerComponent, err := steps.NewSearchControllerComponent(fakeAPIRouter)
	if err != nil {
		fmt.Printf("failed to create search controller component - error: %v", err)
		os.Exit(1)
	}
	apiFeature := controllerComponent.InitAPIFeature()

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()

		err := controllerComponent.Reset()
		if err != nil {
			fmt.Printf("failed to reset controller component - error: %v", err)
			os.Exit(1)
		}
	})

	apiFeature.RegisterSteps(ctx)
	controllerComponent.RegisterSteps(ctx)

	ctx.AfterScenario(func(*godog.Scenario, error) {
		controllerComponent.Close()
		fakeAPIRouter.Close()
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

		if status != 0 {
			t.FailNow()
		}

	} else {
		t.Skip("component flag required to run component tests")
	}
}
