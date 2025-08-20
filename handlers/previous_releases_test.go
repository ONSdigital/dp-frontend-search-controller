package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReadPreviousReleasesWithMigrationLink(t *testing.T) {
	Convey("Given a search handler and zebedee client with a migration link", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		mockedZebedeeClient := &ZebedeeClientMock{
			GetPageDataFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (zebedeeC.PageData, error) {
				return zebedeeC.PageData{
					Type: "bulletin",
					Description: zebedeeC.Description{
						Title:         "My test bulletin",
						Edition:       "March 2024",
						MigrationLink: "/my-new-bulletin",
					},
				}, nil
			},
		}

		mockSearchHandler := NewSearchHandler(&RenderClientMock{}, &SearchClientMock{}, &TopicClientMock{}, mockedZebedeeClient, cfg, cache.List{})

		Convey("When /previousreleases is called", func() {
			req := httptest.NewRequest("GET", "/foo/bar/previousreleases", http.NoBody)

			Convey("Then a 308 redirect should be returned", func() {
				w := doTestRequest("/{uri:.*}/previousreleases", req, mockSearchHandler.PreviousReleases(cfg), nil)
				location := w.Header().Get("Location")
				expectedLocation := "/my-new-bulletin/editions"

				So(w.Code, ShouldEqual, http.StatusPermanentRedirect)
				So(location, ShouldEqual, expectedLocation)
			})
		})
	})
}
