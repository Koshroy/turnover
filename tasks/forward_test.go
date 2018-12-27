package tasks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type mockTransport struct{}

// RoundTrip returns a response in the mock transport for a given request
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "www.example.org" ||
		req.URL.Path != "/inbox" ||
		req.Method != "POST" ||
		req.Header.Get("Content-Type") != "application/ld+json" {
		return nil, fmt.Errorf("should not access URL other than blessed URL")
	}

	body := ioutil.NopCloser(bytes.NewReader([]byte{}))

	header := make(http.Header)
	header.Add("Content-Length", "0")
	header.Add("Content-Type", "application/ld+json")
	header.Add("Date", time.Now().Format(time.RFC1123))

	return &http.Response{
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		ContentLength: 0,
		Request:       req,
		Header:        header,
		Body:          body,
	}, nil
}

func TestForwardTask(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockTransport{},
	}
	task := &Forward{
		TaskID:   "a",
		Activity: []byte{},
		Target: url.URL{
			Scheme:   "https",
			Host:     "www.example.org",
			Path:     "/inbox",
			Fragment: "",
		},
		Client: mockClient,
	}

	if task.ID() != "a" {
		t.Errorf("task ID expected a got %s", task.ID())
		t.FailNow()
	}

	success := task.Run()
	if !success {
		t.Error("task failed to run")
		t.FailNow()
	}
}
