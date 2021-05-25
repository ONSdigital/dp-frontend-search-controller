package service_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx = context.Background()

	errAddCheckFail = errors.New("Error(s) registering checkers for healthcheck")
	errHealthCheck  = errors.New("healthCheck error")
	errServer       = errors.New("HTTP Server error")

	// Health Check Mock
	hcMock = &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
	}
	hcMockAddFail = &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddCheckFail },
		StartFunc:    func(ctx context.Context) {},
	}
	funcDoGetHealthCheckOK = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return hcMock, nil
	}
	funcDoGetHealthCheckFail = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return nil, errHealthCheck
	}
	funcDoGetHealthAddCheckerFail = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return hcMockAddFail, nil
	}

	// Server Mock
	serverWg   = &sync.WaitGroup{}
	serverMock = &mocks.HTTPServerMock{
		ListenAndServeFunc: func() error {
			serverWg.Done()
			return nil
		},
	}
	failingServerMock = &mocks.HTTPServerMock{
		ListenAndServeFunc: func() error {
			serverWg.Done()
			return errServer
		},
	}
	funcDoGetHTTPServerOK = func(bindAddr string, router http.Handler) service.HTTPServer {
		return serverMock
	}
	funcDoGetHTTPServerFail = func(bindAddr string, router http.Handler) service.HTTPServer {
		return failingServerMock
	}

	// Health Client Mock
	funcDoGetHealthClient = func(name string, url string) *health.Client {
		return &health.Client{
			URL:    url,
			Name:   name,
			Client: newMockHTTPClient(&http.Response{}, nil),
		}
	}
)

func TestRunSuccess(t *testing.T) {
	Convey("Given all dependencies are successfully initialised", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckOK,
			DoGetHTTPServerFunc:   funcDoGetHTTPServerOK,
		}
		serverWg.Add(1)
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svcErrors := make(chan error, 1)

			Convey("When Run is called", func() {
				_, err := service.Run(ctx, cfg, mockServiceList, svcErrors)

				Convey("Then service Run is successfully and returns no errors", func() {
					So(err, ShouldBeNil)

					Convey("And the checkers are registered and the healthcheck and http server started", func() {
						So(mockServiceList.HealthCheck, ShouldBeTrue)
						So(len(hcMock.AddCheckCalls()), ShouldEqual, 2)
						So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "frontend renderer")
						So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Search API")
						So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
						So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:25000")
						So(len(hcMock.StartCalls()), ShouldEqual, 1)
						serverWg.Wait() // Wait for HTTP server go-routine to finish
						So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
					})
				})
			})
		})
	})
}

func TestRunFailure(t *testing.T) {
	Convey("Given failure to create healthcheck", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckFail,
		}
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svcErrors := make(chan error, 1)

			Convey("When Run is called", func() {
				_, err := service.Run(ctx, cfg, mockServiceList, svcErrors)

				Convey("Then service Run fails and returns an error", func() {
					So(err, ShouldNotBeNil)
					So(err, ShouldResemble, errHealthCheck)
					So(mockServiceList.HealthCheck, ShouldBeFalse)
				})
			})
		})
	})

	Convey("Given that Checkers cannot be registered", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthAddCheckerFail,
		}
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svcErrors := make(chan error, 1)

			Convey("When Run is called", func() {
				_, err := service.Run(ctx, cfg, mockServiceList, svcErrors)

				Convey("Then service Run fails and returns an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldResemble, errAddCheckFail.Error())

					Convey("And all checks try to register", func() {
						So(mockServiceList.HealthCheck, ShouldBeTrue)
						So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 2)
						So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "frontend renderer")
						So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Search API")
					})
				})
			})
		})
	})

	Convey("Given that HTTP Server fails", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckOK,
			DoGetHTTPServerFunc:   funcDoGetHTTPServerFail,
		}
		serverWg.Add(1)
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svcErrors := make(chan error, 1)

			Convey("When Run is called", func() {
				_, err := service.Run(ctx, cfg, mockServiceList, svcErrors)
				So(err, ShouldBeNil)

				Convey("Then service Run fails and returns an error in the error channel", func() {
					sErr := <-svcErrors
					So(sErr.Error(), ShouldResemble, errServer.Error())
					So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given a correctly initialised service", t, func() {

		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mocks.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mocks.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			serviceList := service.NewServiceList(nil)
			serviceList.HealthCheck = true
			srv := service.Service{
				HealthCheck: hcMock,
				Server:      serverMock,
				ServiceList: serviceList,
			}
			err = srv.Close(ctx, cfg)
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingServerMock := &mocks.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			serviceList := service.NewServiceList(nil)
			serviceList.HealthCheck = true
			srv := service.Service{
				HealthCheck: hcMock,
				Server:      failingServerMock,
				ServiceList: serviceList,
			}
			err = srv.Close(ctx, cfg)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingServerMock.ShutdownCalls()), ShouldEqual, 1)
		})
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string { return []string{} },
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
	}
}
