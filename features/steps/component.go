package steps

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/maxcnunes/httpfake"
)

const (
	gitCommitHash = "3t7e5s1t4272646ef477f8ed755"
	appVersion    = "v1.2.3"
)

// Component contains all the information to create a component test
type Component struct {
	APIFeature      *componentTest.APIFeature
	cfg             *config.Config
	ErrorFeature    componentTest.ErrorFeature
	FakeAPIRouter   *FakeAPI
	FakeRendererApp *FakeAPI
	fakeRequest     *httpfake.Request
	HTTPServer      *http.Server
	ServiceRunning  bool
	svc             *service.Service
	svcErrors       chan error
	StartTime       time.Time
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	c.FakeAPIRouter = NewFakeAPI(&c.ErrorFeature)
	c.FakeRendererApp = NewFakeAPI(&c.ErrorFeature)

	ctx := context.Background()

	svcErrors := make(chan error, 1)

	c.cfg, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.cfg.APIRouterURL = c.FakeAPIRouter.fakeHTTP.ResolveURL("")
	c.cfg.RendererURL = c.FakeRendererApp.fakeHTTP.ResolveURL("")

	c.cfg.HealthCheckInterval = 1 * time.Second
	c.cfg.HealthCheckCriticalTimeout = 2 * time.Second

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:     c.getHTTPServer,
		DoGetHealthCheckFunc:    getHealthCheckOK,
		DoGetHealthClientFunc:   c.getHealthClient,
		DoGetRendererClientFunc: c.getRendererClient,
	}

	serviceList := service.NewServiceList(initFunctions)

	c.svc = service.New()
	if err := c.svc.Init(ctx, c.cfg, serviceList); err != nil {
		log.Event(ctx, "failed to initialise service", log.ERROR, log.Error(err))
		return nil, err
	}

	c.StartTime = time.Now()
	c.svc.Run(ctx, svcErrors)
	c.ServiceRunning = true

	return c, nil
}

// InitAPIFeature initialises the ApiFeature that's contained within a specific JobsFeature.
func (c *Component) InitAPIFeature() *componentTest.APIFeature {
	c.APIFeature = componentTest.NewAPIFeature(c.InitialiseService)

	return c.APIFeature
}

// Reset resets the search controller component
func (c *Component) Reset() *Component {

	c.FakeAPIRouter.Reset()
	c.FakeRendererApp.Reset()
	return c
}

// Close closes the search controller component
func (c *Component) Close() error {
	if c.svc != nil && c.ServiceRunning {
		c.svc.Close(context.Background())
		c.ServiceRunning = false
	}

	c.FakeAPIRouter.Close()
	c.FakeRendererApp.Close()

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

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *Component) getRendererClient(rendererURL string) *renderer.Renderer {
	return &renderer.Renderer{
		HcCli: &health.Client{
			URL:    rendererURL,
			Name:   "renderer",
			Client: c.FakeRendererApp.getMockAPIHTTPClient(),
		},
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
