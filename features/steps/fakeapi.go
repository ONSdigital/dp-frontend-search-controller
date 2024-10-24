package steps

import (
	"strings"

	"github.com/maxcnunes/httpfake"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                *httpfake.HTTPFake
	healthRequest           *httpfake.Request
	searchRequest           *httpfake.Request
	rootTopicRequest        *httpfake.Request
	topicRequest            *httpfake.Request
	subTopicRequest         *httpfake.Request
	subSubTopicRequest      *httpfake.Request
	navigationRequest       *httpfake.Request
	previousReleasesRequest *httpfake.Request
	breadcrumbRequest       *httpfake.Request
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	return &FakeAPI{
		fakeHTTP: httpfake.New(),
	}
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

func (f *FakeAPI) setJSONResponseForGetPageData(url, pageType string, statusCode int) {
	specialCharURL := strings.Replace(url, "/", "%2F", -1)
	path := "/data?uri=" + specialCharURL + "&lang=en"
	bodyStr := `{}`
	if pageType != "" {
		bodyStr = `{"type": "` + pageType + `", "description": {"title": "Labour Market statistics", "edition": "March 2024"}}`
	}
	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).BodyString(bodyStr)
}

func (f *FakeAPI) setJSONResponseForGetBreadcrumb(url string, status int) {
	path := "/parents?uri=" + url
	bodyStr := `[
   		{"uri": "/", "description": {"title": "Home"}, "type": "home_page"}, 
   		{"uri": "/economy", "description": {"title": "Economy"}, "type": "taxonomy_landing_page"},
   		{"uri": "/economy/grossdomesticproductgdp", "description": {"title":"Gross Domestic Product (GDP)"}, "type": "product_page"}
	]`
	f.fakeHTTP.NewHandler().Get(path).Reply(status).BodyString(bodyStr)
}
