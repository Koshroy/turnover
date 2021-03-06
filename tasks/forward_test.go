package tasks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

type mockTransport struct {
	ExpectedReq []byte
}

// RoundTrip returns a response in the mock transport for a given request
// and returns an error if the expected request is not given
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "www.example.org" ||
		req.URL.Path != "/inbox" ||
		req.Method != "POST" ||
		req.Header.Get("Content-Type") != "application/ld+json" {
		return nil, fmt.Errorf("should not access URL other than blessed URL")
	}

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading from body: %v", err)
	}

	if !bytes.Equal(reqBody, m.ExpectedReq) {
		return nil, fmt.Errorf("expected %v got %v", m.ExpectedReq, reqBody)
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
	t.Parallel()

	tID, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error creating taskID: %v", err)
		t.FailNow()
	}

	payload := []byte(`{"key":"value"}`)
	mockClient := &http.Client{
		Transport: &mockTransport{
			ExpectedReq: payload,
		},
	}
	task := &Forward{
		TaskID:   tID,
		Activity: []byte(payload),
		Target: url.URL{
			Scheme:   "https",
			Host:     "www.example.org",
			Path:     "/inbox",
			Fragment: "",
		},
		Client: mockClient,
	}

	if !uuidEqual(task.ID(), tID) {
		t.Errorf("task ID expected %s got %s", tID, task.ID())
		t.FailNow()
	}

	err = task.Run()
	if err != nil {
		t.Errorf("task failed to run, received error: %v", err)
		t.FailNow()
	}
}
