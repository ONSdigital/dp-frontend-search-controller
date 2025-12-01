package steps

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/maxcnunes/httpfake"

	"github.com/cucumber/godog"
)

// RegisterSteps registers the specific steps needed to do component tests for the search controller
func (c *Component) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I wait (\d+) seconds`, c.delayTimeBySeconds)
	ctx.Step(`^all of the downstream services are healthy$`, c.allOfTheDownstreamServicesAreHealthy)
	ctx.Step(`^one of the downstream services is failing`, c.oneOfTheDownstreamServicesIsFailing)
	ctx.Step(`^one of the downstream services is warning`, c.oneOfTheDownstreamServicesIsWarning)
	ctx.Step(`^the page should have the following xml content$`, c.thePageShouldHaveTheFollowingXMLContent)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, c.theResponseHeaderShouldContain)
	ctx.Step(`^the search controller is running$`, c.theSearchControllerIsRunning)
	ctx.Step(`^there is a Search API that gives a successful response and returns ([1-9]\d*|0) results`, c.thereIsASearchAPIThatGivesASuccessfulResponseAndReturnsResults)
	ctx.Step(`^there is a Search API that gives a successful Search URIs response and returns ([1-9]\d*|0) results`, c.thereIsASearchAPIThatGivesASuccessfulSearchURIsResponseAndReturnsResults)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic$`, c.thereIsATopicAPIThatReturnsATopic)
	ctx.Step(`^there is a Topic API returns no topics`, c.thereIsATopicAPIThatReturnsNoTopics)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic and the "([^"]*)" subtopic$`, c.thereIsATopicAPIThatReturnsATopicAndSubtopic)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" topic, the "([^"]*)" subtopic and "([^"]*)" thirdlevel subtopic$`, c.thereIsATopicAPIThatReturnsTheTopicTheSubtopicAndThirdlevelSubtopic)
	ctx.Step(`^there is a Topic API that returns the "([^"]*)" root topic and the "([^"]*)" subtopic for requestQuery "([^"]*)"$`, c.thereIsATopicAPIThatReturnsTheRootTopicAndTheSubtopicForRequestQuery)
	ctx.Step(`^get page data request to zebedee for "([^"]*)" returns a page of type "([^"]*)" with status (\d+)$`, c.getPageDataRequestToZebedeeForReturnsAPageOfTypeWithStatus)
	ctx.Step(`^get page data request to zebedee for "([^"]*)" returns a page with migration link "([^"]*)"$`, c.getPageDataRequestToZebedeeForReturnsAPageWithMigrationLink)
	ctx.Step(`^get page data request to zebedee for "([^"]*)" does not find the page$`, c.getPageDataRequestToZebedeeForDoesNotFindThePage)
	ctx.Step(`^get page data request to zebedee for "([^"]*)" does not have related data`, c.getPageDataRequestToZebedeeThatDoesntContainRelatedData)
	ctx.Step(`^get breadcrumb request to zebedee for "([^"]*)" returns breadcrumbs`, c.getBreadcrumbRequestToZebedeeForReturnsAPageOfTypeWithStatus)
	ctx.Step(`^get breadcrumb request to zebedee for "([^"]*)" fails`, c.getBreadcrumbRequestToZebedeeFails)
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
		log.Error(ctx, "failed to init service", err)
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

func (c *Component) thereIsASearchAPIThatGivesASuccessfulSearchURIsResponseAndReturnsResults(count int) error {
	c.FakeAPIRouter.searchURIsRequest.Lock()
	defer c.FakeAPIRouter.searchURIsRequest.Unlock()

	c.FakeAPIRouter.searchURIsRequest.Response = generateSearchResponse(count)
	return nil
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

	rootTopicAPIResponse := generateRootTopicResponse()
	fakeRootTopicRequestPath := "/topics"
	c.FakeAPIRouter.rootTopicRequest.Get(fakeRootTopicRequestPath)
	c.FakeAPIRouter.rootTopicRequest.Response = rootTopicAPIResponse

	topicAPIResponse := generateTopicResponse(topic)
	fakeTopicRequestPath := fmt.Sprintf("/topics/%s", "6734")
	c.FakeAPIRouter.topicRequest.Get(fakeTopicRequestPath)
	c.FakeAPIRouter.topicRequest.Response = topicAPIResponse

	return nil
}

func (c *Component) thereIsATopicAPIThatReturnsATopicAndSubtopic(topic, subTopic string) error {
	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIResponse := generateRootTopicResponse()
	fakeRootTopicRequestPath := "/topics"
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

func (c *Component) thereIsATopicAPIThatReturnsTheTopicTheSubtopicAndThirdlevelSubtopic(topic, subTopic, subSubTopic string) error {
	c.FakeAPIRouter.rootTopicRequest.Lock()
	defer c.FakeAPIRouter.rootTopicRequest.Unlock()

	c.FakeAPIRouter.topicRequest.Lock()
	defer c.FakeAPIRouter.topicRequest.Unlock()

	rootTopicAPIResponse := generateRootTopicResponse()
	fakeRootTopicRequestPath := "/topics"
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

func (c *Component) thePageShouldHaveTheFollowingXMLContent(body *godog.DocString) error {
	tmpExpected := string(c.FakeAPIRouter.subTopicRequest.Response.BodyBuffer)
	actual := strings.Replace(strings.Replace(strings.TrimSpace(tmpExpected), "\n", "", -1), "\t", "", -1)
	actual = strings.Join(strings.Fields(strings.TrimSpace(actual)), " ")
	actual = strings.Replace(actual, "><", "> <", -1)

	expected := strings.Replace(strings.Replace(strings.TrimSpace(body.Content), "\n", "", -1), "\t", "", -1)
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

func (c *Component) getPageDataRequestToZebedeeForReturnsAPageOfTypeWithStatus(url, pageType string, statusCode int) error {
	c.FakeAPIRouter.setJSONResponseForGetPageData(url, pageType, statusCode, true)
	return nil
}

func (c *Component) getPageDataRequestToZebedeeForReturnsAPageWithMigrationLink(url, migrationLink string) error {
	c.FakeAPIRouter.pageDataRequest.Lock()
	defer c.FakeAPIRouter.pageDataRequest.Unlock()

	specialCharURL := strings.Replace(url, "/", "%2F", -1)
	path := "/data?uri=" + specialCharURL + "&lang=en"

	c.FakeAPIRouter.pageDataRequest.Get(path)
	c.FakeAPIRouter.pageDataRequest.Response = generatePageDataWithMigrationLink(migrationLink)

	return nil
}

func (c *Component) getPageDataRequestToZebedeeForDoesNotFindThePage(url string) error {
	c.FakeAPIRouter.setJSONResponseForGetPageData(url, "", 404, false)
	return nil
}

func (c *Component) getPageDataRequestToZebedeeThatDoesntContainRelatedData(url string) error {
	c.FakeAPIRouter.setJSONResponseForGetPageData(url, "bulletin", 200, false)
	return nil
}

func (c *Component) getBreadcrumbRequestToZebedeeForReturnsAPageOfTypeWithStatus(url string) error {
	c.FakeAPIRouter.setJSONResponseForGetBreadcrumb(url, 200)
	return nil
}

func (c *Component) getBreadcrumbRequestToZebedeeFails(url string) error {
	c.FakeAPIRouter.setJSONResponseForGetBreadcrumb(url, 500)
	return nil
}
