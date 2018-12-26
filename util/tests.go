package util

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type MockTransport struct {
	Fallback http.RoundTripper
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "www.w3.org" && req.URL.Path != "/ns/activitystreams" {
		return m.Fallback.RoundTrip(req)
	}

	f, err := os.Open("./testdata/activitystreams.jsonld")
	if err != nil {
		return nil, fmt.Errorf("error opening testdata for mock transport: %v", err)
	}

	s, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting filesize for mock transport: %v", err)
	}

	header := make(http.Header)
	header.Add("Content-Length", fmt.Sprintf("%d", s.Size()))
	header.Add("Content-Type", "application/ld+json")
	header.Add("Date", s.ModTime().Format(time.RFC1123))

	return &http.Response{
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		ContentLength: s.Size(),
		Request:       req,
		Header:        header,
		Body:          f,
	}, nil
}
