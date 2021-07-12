package steps

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/maxcnunes/httpfake"
)

// Component contains all the information to create a component test
type Component struct {
	componenttest.ErrorFeature
	FakeAPIRouter   *FakeAPI
	FakeRendererApp *FakeAPI
	fakeRequest     *httpfake.Request
	HTTPServer      *http.Server
	ServiceRunning  bool
	svc             *service.Service
	svcErrors       chan error
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (c *Component, err error) {
	c = &Component{
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	c.FakeAPIRouter = NewFakeAPI(c)
	c.FakeRendererApp = NewFakeAPI(c)

	ctx := context.Background()

	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	svcErrors := make(chan error, 1)

	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	cfg.APIRouterURL = c.FakeAPIRouter.fakeHTTP.ResolveURL("")
	cfg.RendererURL = c.FakeRendererApp.fakeHTTP.ResolveURL("")

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:     c.getHTTPServer,
		DoGetHealthCheckFunc:    getHealthCheckOK,
		DoGetHealthClientFunc:   c.getHealthClient,
		DoGetRendererClientFunc: c.getRendererClient,
	}

	serviceList := service.NewServiceList(initFunctions)

	c.svc = service.New()
	if err := c.svc.Init(ctx, cfg, serviceList); err != nil {
		log.Event(ctx, "failed to initialise service", log.ERROR, log.Error(err))
		return nil, err
	}

	c.svc.Run(ctx, svcErrors)
	c.ServiceRunning = true

	return c, nil
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
	return &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		HandlerFunc:  func(w http.ResponseWriter, req *http.Request) {},
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
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
