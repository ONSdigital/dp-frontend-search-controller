package steps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	zebedeeCli "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/handlers"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	searchModels "github.com/ONSdigital/dp-search-api/models"
	topicModels "github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/maxcnunes/httpfake"
)

const (
	gitCommitHash = "3t7e5s1t4272646ef477f8ed755"
	appVersion    = "v1.2.3"
)

// Component contains all the information to create a component test
type Component struct {
	APIFeature     *componentTest.APIFeature
	Config         *config.Config
	ErrorFeature   componentTest.ErrorFeature
	FakeAPIRouter  *FakeAPI
	fakeRequest    *httpfake.Request
	HTTPServer     *http.Server
	ServiceRunning bool
	svc            *service.Service
	svcErrors      chan error
	StartTime      time.Time
	zebedeeClient  handlers.ZebedeeClient
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	ctx := context.Background()

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.Config.EnableAggregationPages = true
	c.Config.EnableTopicAggregationPages = true

	log.Info(ctx, "configuration for component test", log.Data{"config": c.Config})

	c.FakeAPIRouter = NewFakeAPI()
	c.Config.APIRouterURL = c.FakeAPIRouter.fakeHTTP.ResolveURL("")

	c.Config.HealthCheckInterval = 1 * time.Second
	c.Config.HealthCheckCriticalTimeout = 3 * time.Second

	c.FakeAPIRouter.healthRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/health")
	c.FakeAPIRouter.healthRequest.CustomHandle = healthCheckStatusHandle(200)

	c.FakeAPIRouter.searchRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/search")
	c.FakeAPIRouter.searchRequest.Response = generateSearchResponse(1)

	c.FakeAPIRouter.rootTopicRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get(fmt.Sprintf("/topics"))
	c.FakeAPIRouter.topicRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/topics/*")
	c.FakeAPIRouter.subTopicRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/topics/*")
	c.FakeAPIRouter.subSubTopicRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/topics/*")
	c.FakeAPIRouter.previousReleasesRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/economy/previousreleases")
	c.FakeAPIRouter.navigationRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/data")
	c.FakeAPIRouter.breadcrumbRequest = c.FakeAPIRouter.fakeHTTP.NewHandler().Get("/parents")

	c.zebedeeClient = zebedeeCli.New(c.Config.APIRouterURL)

	// Please use the step to start the service - this is down to
	// the auto updates against backing services are hard to predict so
	// it is easier to provision them first and then start the service.
	return c, nil
}

// InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (c *Component) InitAPIFeature() *componentTest.APIFeature {
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.APIFeature
}

// Close closes the search controller component
func (c *Component) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}

	c.FakeAPIRouter.Close()

	return nil
}

// InitialiseService returns the http.Handler that's contained within the component.
func (c *Component) InitialiseService() (http.Handler, error) {
	return c.HTTPServer.Handler, nil
}

func getHealthCheckOK(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	componentBuildTime := strconv.Itoa(int(time.Now().Unix()))
	versionInfo, err := healthcheck.NewVersionInfo(componentBuildTime, gitCommitHash, appVersion)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (c *Component) getHealthClient(name string, url string) *health.Client {
	return &health.Client{
		URL:    url,
		Name:   name,
		Client: c.FakeAPIRouter.getMockAPIHTTPClient(),
	}
}

// newMock mocks HTTP Client
func (f *FakeAPI) getMockAPIHTTPClient() *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string { return []string{} },
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return f.fakeHTTP.Server.Client().Do(req)
		},
	}
}

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func generateSearchResponse(count int) *httpfake.Response {
	searchAPIResponse := searchModels.SearchResponse{
		Count: count,
		Items: []searchModels.Item{},
	}

	for i := 0; i < count; i++ {
		newSearchItem := generateSearchItem(i)
		searchAPIResponse.Items = append(searchAPIResponse.Items, newSearchItem)
	}

	fakeAPIResponse := httpfake.NewResponse()
	fakeAPIResponse.Status(200)
	fakeAPIResponse.BodyStruct(searchAPIResponse)

	return fakeAPIResponse
}

func generateSearchItem(num int) searchModels.Item {

	title := fmt.Sprintf("Test Bulletin %d", num)
	uri := fmt.Sprintf("http://localhost://test-bulletin-%d", num)
	cdid := fmt.Sprintf("AA%d", num)
	datasetID := fmt.Sprintf("DD%d", num)

	searchItem := searchModels.Item{
		Title:     title,
		URI:       uri,
		CDID:      cdid,
		DatasetID: datasetID,
	}
	return searchItem
}

func generateRootTopicResponse() *httpfake.Response {

	var (
		testEconomyRootTopic = topicModels.Topic{
			ID:          "6734",
			Title:       "Economy",
			SubtopicIds: &[]string{"1834"},
		}

		testBusinessRootTopic = topicModels.Topic{
			ID:          "1234",
			Title:       "Business",
			SubtopicIds: &[]string{},
		}
	)

	rootTopicAPIResponse := &topicModels.PublicSubtopics{
		Count:       2,
		Offset:      0,
		Limit:       50,
		TotalCount:  2,
		PublicItems: &[]topicModels.Topic{testEconomyRootTopic, testBusinessRootTopic},
	}

	fakeAPIResponse := httpfake.NewResponse()
	fakeAPIResponse.Status(200)
	fakeAPIResponse.BodyStruct(rootTopicAPIResponse)

	return fakeAPIResponse
}

func generateTopicResponse(title string) *httpfake.Response {
	topicAPIResponse := &topicModels.Topic{}
	switch title {
	case "root":
		topicAPIResponse = &topicModels.Topic{
			ID:          "9999",
			Slug:        title,
			Title:       "Root",
			SubtopicIds: &[]string{"6734", "1234"},
		}
	case "economy":
		topicAPIResponse = &topicModels.Topic{
			ID:          "6734",
			Slug:        title,
			Title:       "Economy",
			SubtopicIds: &[]string{"1834", "8286"},
		}
	case "environmentalaccounts":
		topicAPIResponse = &topicModels.Topic{
			ID:          "1834",
			Slug:        title,
			Title:       "Environmental Accounts",
			SubtopicIds: &[]string{},
		}
	case "governmentpublicsectorandtaxes":
		topicAPIResponse = &topicModels.Topic{
			ID:          "8286",
			Slug:        title,
			Title:       "Government, public sector and taxes",
			SubtopicIds: &[]string{"3687"},
		}
	case "publicsectorfinance":
		topicAPIResponse = &topicModels.Topic{
			ID:          "3687",
			Slug:        title,
			Title:       "Public sector finance",
			SubtopicIds: &[]string{},
		}
	case "business":
		topicAPIResponse = &topicModels.Topic{
			ID:          "1234",
			Slug:        title,
			Title:       "Business",
			SubtopicIds: &[]string{},
		}
	}

	fakeAPIResponse := httpfake.NewResponse()
	fakeAPIResponse.Status(200)
	fakeAPIResponse.BodyStruct(topicAPIResponse)

	return fakeAPIResponse
}

func generateEmptyTopicResponse() *httpfake.Response {
	topicAPIResponse := &topicModels.Topic{}

	fakeAPIResponse := httpfake.NewResponse()
	fakeAPIResponse.Status(200)
	fakeAPIResponse.BodyStruct(topicAPIResponse)

	return fakeAPIResponse
}

func generateTopicResponseRSS(topic, subtopic string) *httpfake.Response {
	topicTitle := "Latest ONS releases matching"
	subTopicTitle := "Latest ONS releases"
	query := "rss"
	topicSubTopicRSSResponse := `
		<?xml version="1.0" encoding="ISO-8859-1"?><rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
		<channel>
		<title>` + topicTitle + ` (topic: /` + topic + `/` + subtopic + `, type: data).</title>
		<link>http://dp.aws.onsdigital.uk/` + topic + `/` + subtopic + `?` + query + `</link>
		<description>` + subTopicTitle + `</description> <category>/` + topic + `/` + subtopic + `</category>
		<dc:subject>/` + topic + `/` + subtopic + `</dc:subject></channel>
		</rss>`

	fakeAPIResponse := httpfake.NewResponse()
	fakeAPIResponse.Header.Add("Content-Type", "application/rss+xml; charset=ISO-8859-1")
	fakeAPIResponse.Status(200)
	fakeAPIResponse.BodyString(topicSubTopicRSSResponse)

	return fakeAPIResponse
}
