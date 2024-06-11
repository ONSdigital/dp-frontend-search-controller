package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/log.go/v2/log"
	"io/ioutil"
	"net/http"
	"strings"
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
	ctx.Step(`^I should receive the following health JSON response:$`, c.iShouldReceiveTheFollowingHealthJSONResponse)
	ctx.Step(`^I wait (\d+) seconds`, c.delayTimeBySeconds)
	ctx.Step(`^all of the downstream services are healthy$`, c.allOfTheDownstreamServicesAreHealthy)
	ctx.Step(`^one of the downstream services is failing`, c.oneOfTheDownstreamServicesIsFailing)
	ctx.Step(`^one of the downstream services is warning`, c.oneOfTheDownstreamServicesIsWarning)
	ctx.Step(`^the page should have the following xml content$`, c.thePageShouldHaveTheFollowingXmlContent)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, c.theResponseHeaderShouldContain)
	ctx.Step(`^the search controller is running$`, c.theSearchControllerIsRunning)
	ctx.Step(`^there is a Search API that gives a successful response and returns ([1-9]\d*|0) results`, c.thereIsASearchAPIThatGivesASuccessfulResponseAndReturnsResults)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic$`, c.thereIsATopicAPIThatReturnsATopic)
	ctx.Step(`^there is a Topic API returns no topics`, c.thereIsATopicAPIThatReturnsNoTopics)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic and the "([^"]*)" subtopic$`, c.thereIsATopicAPIThatReturnsATopicAndSubtopic)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic, the "([^"]*)" subtopic and "([^"]*)" thirdlevel subtopic$`, c.thereIsATopicAPIThatReturnsTheTopicTheSubtopicAndThirdlevelSubtopic)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" root topic and the "([^"]*)" subtopic for requestQuery "([^"]*)"$`, c.thereIsATopicAPIThatReturnsTheRootTopicAndTheSubtopicForRequestQuery)

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

func (c *Component) thereIsATopicAPIThatReturnsATopic(topic string) error {

	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIResponse := generateTopicResponse("root")
	fakeRootTopicRequestPath := fmt.Sprintf("/topics/%s", "9999")
	c.FakeAPIRouter.rootTopicRequest.Get(fakeRootTopicRequestPath)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIResponse

	topicAPIResponse := generateTopicResponse(topic)
	fakeTopicRequestPath := fmt.Sprintf("/topics/%s", "6734")
	c.FakeAPIRouter.topicRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsATopicAndSubtopic(topic string, subTopic string) error {

	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIResponse := generateTopicResponse("root")
	fakeRootTopicRequestPath := fmt.Sprintf("/topics/%s", "9999")
	c.FakeAPIRouter.rootTopicRequest.Get(fakeRootTopicRequestPath)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIResponse

	topicAPIResponse := generateTopicResponse(topic)
	fakeTopicRequestPath := fmt.Sprintf("/topics/%s", "6734")
	c.FakeAPIRouter.topicRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	subTopicAPIResponse := generateTopicResponse(subTopic)
	fakeSubTopicRequestPath := fmt.Sprintf("/topics/%s", "1834")
	c.FakeAPIRouter.subTopicRequest.Get(fakeSubTopicRequestPath)
	c.FakeAPIRouter.subTopicRequest.Response = subTopicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsTheTopicTheSubtopicAndThirdlevelSubtopic(topic string, subTopic string, subSubTopic string) error {

	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIResponse := generateTopicResponse("root")
	fakeRootTopicRequestPath := fmt.Sprintf("/topics/%s", "9999")
	c.FakeAPIRouter.rootTopicRequest.Get(fakeRootTopicRequestPath)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIResponse

	topicAPIResponse := generateTopicResponse(topic)
	fakeTopicRequestPath := fmt.Sprintf("/topics/%s", "6734")
	c.FakeAPIRouter.topicRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	subTopicAPIResponse := generateTopicResponse(subTopic)
	fakeSubTopicRequestPath := fmt.Sprintf("/topics/%s", "8286")
	c.FakeAPIRouter.subTopicRequest.Get(fakeSubTopicRequestPath)
	c.FakeAPIRouter.subTopicRequest.Response = subTopicAPIResponse

	subSubTopicAPIResponse := generateTopicResponse(subSubTopic)
	fakeSubSubTopicRequestPath := fmt.Sprintf("/topics/%s", "3687")
	c.FakeAPIRouter.subSubTopicRequest.Get(fakeSubSubTopicRequestPath)
	c.FakeAPIRouter.subSubTopicRequest.Response = subSubTopicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsNoTopics() error {

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	topicAPIResponse := generateEmptyTopicResponse()
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	return nil
}

func (c *Component) thePageShouldHaveTheFollowingXmlContent(body *godog.DocString) error {

	tmpExpected := string(c.FakeAPIRouter.subTopicRequest.Response.BodyBuffer[:])
	actual := strings.Replace(strings.Replace(strings.TrimSpace(string(tmpExpected[:])), "\n", "", -1), "\t", "", -1)
	actual = strings.Join(strings.Fields(strings.TrimSpace(actual)), " ")
	actual = strings.Replace(actual, "><", "> <", -1)

	expected := strings.Replace(strings.Replace(strings.TrimSpace(string(body.Content[:])), "\n", "", -1), "\t", "", -1)
	expected = strings.Join(strings.Fields(strings.TrimSpace(expected)), " ")
	expected = strings.Replace(expected, "><", "> <", -1)

	if actual != expected {
		return errors.New("expected body to be: " + "\n" + expected + "\n\t but actual is: " + "\n" + actual)
	}
	return nil
}

func (c *Component) theResponseHeaderShouldContain(key, value string) (err error) {
	responseHeader := c.FakeAPIRouter.subTopicRequest.Response.Header
	actualValue, actualExist := responseHeader[key]
	if !actualExist {
		return errors.New("expected header key " + key + ", does not exist in the header ")
	}
	if actualValue[0] != value {
		return errors.New("expected header value " + value + ", but is actually is :" + actualValue[0])
	}

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsTheRootTopicAndTheSubtopicForRequestQuery(topic, subTopic, query string) error {

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	topicAPIResponse := generateTopicResponseRSS(topic, subTopic)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	fakeTopicRequestPath := fmt.Sprintf("/%s/%s?%s", topic, subTopic, query)

	c.FakeAPIRouter.subTopicRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.subTopicRequest.Response = topicAPIResponse

	return nil
}
