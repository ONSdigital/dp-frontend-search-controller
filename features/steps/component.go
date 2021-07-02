package steps

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/maxcnunes/httpfake"
)

// Component contains all the information to create a component test
type Component struct {
	componenttest.ErrorFeature
	cfg             *config.Config
	ctx             context.Context
	FakeAPIRouter   *FakeAPI
	FakeRendererApp *FakeAPI
	fakeRequest     *httpfake.Request
	HTTPServer      *http.Server
	serviceList     *service.ExternalServiceList
	ServiceRunning  bool
	signals         chan os.Signal
	svc             *service.Service
	svcErrors       chan error
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (c *Component, err error) {
	c = &Component{
		ctx:        context.Background(),
		HTTPServer: &http.Server{},
		svcErrors:  make(chan error),
	}

	c.FakeAPIRouter = NewFakeAPI(c)
	c.FakeRendererApp = NewFakeAPI(c)

	c.signals = make(chan os.Signal, 1)
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)

	c.svcErrors = make(chan error, 1)

	c.cfg, err = config.Get()
	if err != nil {
		return nil, err
	}

	c.cfg.APIRouterURL = c.FakeAPIRouter.fakeHTTP.ResolveURL("")
	c.cfg.RendererURL = c.FakeRendererApp.fakeHTTP.ResolveURL("")

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:     c.getHTTPServer,
		DoGetHealthCheckFunc:    getHealthCheckOK,
		DoGetHealthClientFunc:   getHealthClient,
		DoGetRendererClientFunc: getRendererClient,
	}

	c.serviceList = service.NewServiceList(initFunctions)

	c.svc = service.New()

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

// InitialiseService initialises search controller component service
func (c *Component) InitialiseService() (http.Handler, error) {
	if err := c.svc.Init(c.ctx, c.cfg, c.serviceList); err != nil {
		log.Event(c.ctx, "failed to initialise service", log.ERROR, log.Error(err))
		return nil, err
	}

	c.svc.Run(c.ctx, c.svcErrors)

	c.ServiceRunning = true

	return c.HTTPServer.Handler, nil
}

func getHealthCheckOK(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func getHealthClient(name string, url string) *health.Client {
	return &health.Client{
		URL:    url,
		Name:   name,
		Client: service.NewMockHTTPClient(&http.Response{}, nil),
	}
}

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func getRendererClient(rendererURL string) *renderer.Renderer {
	return &renderer.Renderer{
		HcCli: &health.Client{
			URL:    rendererURL,
			Name:   "renderer",
			Client: service.NewMockHTTPClient(&http.Response{}, nil),
		},
	}
}
