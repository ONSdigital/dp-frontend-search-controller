package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/log.go/v2/log"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/assert"
)

// HealthCheckTest represents a test healthcheck struct that mimics the real healthcheck struct
type HealthCheckTest struct {
	Status    string                  `json:"status"`
	Version   healthcheck.VersionInfo `json:"version"`
	Uptime    time.Duration           `json:"uptime"`
	StartTime time.Time               `json:"start_time"`
	Checks    []*Check                `json:"checks"`
}

// Check represents a health status of a registered app that mimics the real check struct
// As the component test needs to access fields that are not exported in the real struct
type Check struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StatusCode  int        `json:"status_code"`
	Message     string     `json:"message"`
	LastChecked *time.Time `json:"last_checked"`
	LastSuccess *time.Time `json:"last_success"`
	LastFailure *time.Time `json:"last_failure"`
}

// RegisterSteps registers the specific steps needed to do component tests for the search controller
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the search controller is running$`, c.theSearchControllerIsRunning)
	ctx.Step(`^I wait (\d+) seconds`, c.delayTimeBySeconds)
	ctx.Step(`^all of the downstream services are healthy$`, c.allOfTheDownstreamServicesAreHealthy)
	ctx.Step(`^one of the downstream services is warning`, c.oneOfTheDownstreamServicesIsWarning)
	ctx.Step(`^one of the downstream services is failing`, c.oneOfTheDownstreamServicesIsFailing)
	ctx.Step(`^I should receive the following health JSON response:$`, c.iShouldReceiveTheFollowingHealthJSONResponse)
	ctx.Step(`^there is a Search API that gives a successful response and returns ([1-9]\d*|0) results`, c.thereIsASearchAPIThatGivesASuccessfulResponseAndReturnsResults)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic$`, c.thereIsATopicAPIThatReturnsARootTopic)
	ctx.Step(`^there is a Topic API returns no topics`, c.thereIsATopicAPIThatReturnsNoTopics)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic and the "([^"]*)" subtopic$`, c.thereIsATopicAPIThatReturnsATopicAndSubtopic)
}

func (c *Component) theSearchControllerIsRunning() error {
	ctx := context.Background()

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:   c.getHTTPServer,
		DoGetHealthCheckFunc:  getHealthCheckOK,
		DoGetHealthClientFunc: c.getHealthClient,
	}

	serviceList := service.NewServiceList(initFunctions)

	c.svc = service.New()
	if err := c.svc.Init(ctx, c.Config, serviceList); err != nil {
		log.Error(ctx, "failed to initialise service", err)
		return err
	}

	svcErrors := make(chan error, 1)

	c.StartTime = time.Now()
	c.svc.Run(ctx, svcErrors)
	c.ServiceRunning = true

	return nil
}

// delayTimeBySeconds pauses the goroutine for the given seconds
func (c *Component) delayTimeBySeconds(sec int) error {
	time.Sleep(time.Duration(int64(sec)) * time.Second)
	return nil
}

func (c *Component) allOfTheDownstreamServicesAreHealthy() error {
	c.FakeAPIRouter.healthRequest.Lock()
	defer c.FakeAPIRouter.healthRequest.Unlock()

	c.FakeAPIRouter.healthRequest.CustomHandle = healthCheckStatusHandle(200)

	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsWarning() error {
	c.FakeAPIRouter.healthRequest.Lock()
	defer c.FakeAPIRouter.healthRequest.Unlock()

	c.FakeAPIRouter.healthRequest.CustomHandle = healthCheckStatusHandle(429)

	return nil
}

func (c *Component) oneOfTheDownstreamServicesIsFailing() error {
	c.FakeAPIRouter.healthRequest.Lock()
	defer c.FakeAPIRouter.healthRequest.Unlock()

	c.FakeAPIRouter.healthRequest.CustomHandle = healthCheckStatusHandle(500)

	return nil
}

func healthCheckStatusHandle(status int) httpfake.Responder {
	return func(w http.ResponseWriter, r *http.Request, rh *httpfake.Request) {
		rh.Lock()
		defer rh.Unlock()
		w.WriteHeader(status)
	}
}

func (c *Component) iShouldReceiveTheFollowingHealthJSONResponse(expectedResponse *godog.DocString) error {
	var healthResponse, expectedHealth HealthCheckTest

	responseBody, err := ioutil.ReadAll(c.APIFeature.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response of search controller component - error: %v", err)
	}

	err = json.Unmarshal(responseBody, &healthResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response of search controller component - error: %v", err)
	}

	err = json.Unmarshal([]byte(expectedResponse.Content), &expectedHealth)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expected health response - error: %v", err)
	}

	c.validateHealthCheckResponse(healthResponse, expectedHealth)

	return c.ErrorFeature.StepError()
}

func (c *Component) validateHealthCheckResponse(healthResponse HealthCheckTest, expectedResponse HealthCheckTest) {
	maxExpectedStartTime := c.StartTime.Add((c.Config.HealthCheckInterval + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedResponse.Status, healthResponse.Status)
	assert.True(&c.ErrorFeature, healthResponse.StartTime.After(c.StartTime))
	assert.True(&c.ErrorFeature, healthResponse.StartTime.Before(maxExpectedStartTime))
	assert.Greater(&c.ErrorFeature, healthResponse.Uptime.Seconds(), float64(0))

	c.validateHealthVersion(healthResponse.Version, expectedResponse.Version, maxExpectedStartTime)

	for i, checkResponse := range healthResponse.Checks {
		c.validateHealthCheck(checkResponse, expectedResponse.Checks[i])
	}
}

func (c *Component) validateHealthVersion(versionResponse healthcheck.VersionInfo, expectedVersion healthcheck.VersionInfo, maxExpectedStartTime time.Time) {
	assert.True(&c.ErrorFeature, versionResponse.BuildTime.Before(maxExpectedStartTime))
	assert.Equal(&c.ErrorFeature, expectedVersion.GitCommit, versionResponse.GitCommit)
	assert.Equal(&c.ErrorFeature, expectedVersion.Language, versionResponse.Language)
	assert.NotEmpty(&c.ErrorFeature, versionResponse.LanguageVersion)
	assert.Equal(&c.ErrorFeature, expectedVersion.Version, versionResponse.Version)
}

func (c *Component) validateHealthCheck(checkResponse, expectedCheck *Check) {
	maxExpectedHealthCheckTime := c.StartTime.Add((c.Config.HealthCheckInterval + c.Config.HealthCheckCriticalTimeout + 1) * time.Second)

	assert.Equal(&c.ErrorFeature, expectedCheck.Name, checkResponse.Name)
	assert.Equal(&c.ErrorFeature, expectedCheck.Status, checkResponse.Status)
	assert.Equal(&c.ErrorFeature, expectedCheck.StatusCode, checkResponse.StatusCode)
	assert.Equal(&c.ErrorFeature, expectedCheck.Message, checkResponse.Message)
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.Before(maxExpectedHealthCheckTime))
	assert.True(&c.ErrorFeature, checkResponse.LastChecked.After(c.StartTime))

	if expectedCheck.StatusCode == 200 {
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastSuccess.After(c.StartTime))
	} else {
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.Before(maxExpectedHealthCheckTime))
		assert.True(&c.ErrorFeature, checkResponse.LastFailure.After(c.StartTime))
	}
}

func (c *Component) thereIsASearchAPIThatGivesASuccessfulResponseAndReturnsResults(count int) error {
	c.FakeAPIRouter.searchRequest.Lock()
	defer c.FakeAPIRouter.searchRequest.Unlock()

	c.FakeAPIRouter.searchRequest.Response = generateSearchResponse(count)

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsARootTopic(topic string) error {

	// This should be configurable really
	fakeTopicId := "6734"

	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIRespose := generateTopicResponseWithSubtopic(fakeTopicId, topic)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIRespose

	topicAPIResponse := generateTopicResponse(fakeTopicId, topic)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsATopicAndSubtopic(topic string, subTopic string) error {

	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	economyTopicID := "6734"
	SubTopicTitle := "environmental accounts"

	rootTopicAPIRespose := generateTopicResponseWithSubtopic(economyTopicID, topic)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIRespose

	topicAPIResponse := generateTopicResponse(economyTopicID, subTopic)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	subtopicAPIResponse := generateTopicResponseWithSubtopic("1834", SubTopicTitle)

	fakeTopicRequestPath := fmt.Sprintf("/topics/%s/subtopics", economyTopicID)
	c.FakeAPIRouter.subtopicsRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.subtopicsRequest.Response = subtopicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsNoTopics() error {

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	topicAPIResponse := generateEmptyTopicResponse()
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	return nil
}
