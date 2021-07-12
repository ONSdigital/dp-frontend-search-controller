package steps

import (
	"github.com/cucumber/godog"
)

// RegisterSteps registers the specific steps needed to do component tests for the search controller
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^all the downstream services are healthy$`, c.allTheDownstreamServicesAreHealthy)
	ctx.Step(`^one the downstream services is warning`, c.oneOfTheDownstreamServicesIsWarning)
	ctx.Step(`^one the downstream services is failing`, c.oneOfTheDownstreamServicesIsFailing)
}

func (c *Component) allTheDownstreamServicesAreHealthy() error {
	c.FakeAPIRouter.setJSONResponseForGet("/health", 500)
	c.FakeRendererApp.setJSONResponseForGet("/health", 500)
	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsWarning() error {
	c.FakeAPIRouter.setJSONResponseForGet("/health", 429)
	c.FakeRendererApp.setJSONResponseForGet("/health", 200)
	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsFailing() error {
	c.FakeAPIRouter.setJSONResponseForGet("/health", 500)
	c.FakeRendererApp.setJSONResponseForGet("/health", 200)
	return nil
}
