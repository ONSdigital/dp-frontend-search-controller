package steps

import (
	"github.com/maxcnunes/httpfake"
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

func (f *FakeAPI) setJSONResponseForGetPageData(statusCode int) {
	if statusCode == 200 {
		f.fakeHTTP.NewHandler().Get("/data?uri=%2Feconomy%2Flatest&lang=en").Reply(statusCode).BodyString(`{"type": "article",
									"description": {"title": "labour market statistics"}}`)
	} else {
		f.fakeHTTP.NewHandler().Get("/data?uri=%2Feconomy%2Flatest&lang=en").Reply(statusCode).BodyString(`{"type": "taxonomy_landing_page",
									"description": {"title": "Economic output and productivity"}}`)
	}
}
