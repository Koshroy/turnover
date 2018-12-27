package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const followJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": "https://activities.example.org/1",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.com/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const emptyIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": "",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.com/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const nullIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": null,
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.com/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const missingIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.com/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const noteJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Note",
    "id": "https://activities.example.org/2",
    "actor": "https://sally.example.org",
    "object": {
        "id": "https://notes.example.com/1"
    }
}
`

const createNoteJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Create",
    "id": "https://activities.example.org/3",
    "actor": "https://sally.otherexample.org",
    "object": {
        "summary": "Note",
        "type": "Note",
        "id": "https://sally.otherexample.org/note/1",
        "attributedTo": "https://john.otherexample.org"
    }
}
`

type mockTransport struct {
	Fallback http.RoundTripper
}

// RoundTrip returns a response in the mock transport for a given request
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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

type respTest struct {
	JSONInput  string
	StatusCode int
	TestName   string
}

func testResp(t *testing.T, inbox *Inbox, tests []respTest) {
	for _, tt := range tests {
		t.Run("request_status_code_"+tt.TestName, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.JSONInput))
			w := httptest.NewRecorder()
			inbox.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.StatusCode {
				t.Errorf("expected %d got %d", tt.StatusCode, resp.StatusCode)
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					t.Logf("response body: %v\n", string(respBytes))
				}
				t.FailNow()
			}

		})
	}
}

func TestInboxHandler(t *testing.T) {
	t.Parallel()

	mockClient := &http.Client{
		Transport: &mockTransport{Fallback: http.DefaultTransport},
	}
	i := NewInbox([]string{}, "https", "www.example.com", mockClient)

	testResp(t, i, []respTest{
		{followJSON, http.StatusOK, "success_follow_json"},
		{emptyIDFollowJSON, http.StatusOK, "success_follow_json_empty_id"},
		{nullIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_null_id"},
		{missingIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_missing_id"},
		{noteJSON, http.StatusUnsupportedMediaType, "failure_note_json"},
		{createNoteJSON, http.StatusOK, "success_create_note_json"},
	})

}
