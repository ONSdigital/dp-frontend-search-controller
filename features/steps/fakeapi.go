package steps

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/maxcnunes/httpfake"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                     *httpfake.HTTPFake
	healthRequest                *httpfake.Request
	outboundRequests             []string
	collectOutboundRequestBodies httpfake.CustomAssertor
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	fa := &FakeAPI{
		fakeHTTP: httpfake.New(),
	}

	fa.collectOutboundRequestBodies = func(r *http.Request) error {
		// inspect request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading the outbound request body: %s", err.Error())
		}
		fa.outboundRequests = append(fa.outboundRequests, string(body))
		return nil
	}

	return fa
}

func (f *FakeAPI) setJSONResponseForGet(url string, statusCode int) {
	f.fakeHTTP.NewHandler().Get(url).Reply(statusCode)
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

// Reset resets the fake API
func (f *FakeAPI) Reset() {
	f.fakeHTTP.Reset()
}
