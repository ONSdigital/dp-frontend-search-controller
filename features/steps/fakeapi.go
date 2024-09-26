package steps

import (
	"github.com/maxcnunes/httpfake"
	"strings"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                     *httpfake.HTTPFake
	healthRequest                *httpfake.Request
	searchRequest                *httpfake.Request
	rootTopicRequest             *httpfake.Request
	topicRequest                 *httpfake.Request
	subTopicRequest              *httpfake.Request
	subSubTopicRequest           *httpfake.Request
	navigationRequest            *httpfake.Request
	zebedeeRequest               *httpfake.Request
	previousReleasesRequest      *httpfake.Request
	outboundRequests             []string
	collectOutboundRequestBodies httpfake.CustomAssertor
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

func (f *FakeAPI) setJSONResponseForGetPageData(url string, pageType string, statusCode int) {
	specialCharUrl := strings.Replace(url, "/", "%2F", -1)
	path := "/data?uri=" + specialCharUrl + "&lang=en"
	bodyStr := `{}`
	if pageType != "" {
		bodyStr = `{"type": "` + pageType + `", "description": {"title": "labour market statistics"}}`
	}
	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).BodyString(bodyStr)
}
