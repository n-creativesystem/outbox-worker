package mock

import "net/http"

type MockRoundTripper func(*http.Request) (*http.Response, error)

func (rt MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}
