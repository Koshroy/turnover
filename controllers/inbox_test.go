package controllers

import (
	// "encoding/json"
	// "io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const followJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": "https://activities.example.org/1",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "http://www.example.com/inbox",
        "attributedTo": "http://john.example.org",
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
        "id": "http://www.example.com/inbox",
        "attributedTo": "http://john.example.org",
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
        "id": "http://www.example.com/inbox",
        "attributedTo": "http://john.example.org",
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
        "id": "http://www.example.com/inbox",
        "attributedTo": "http://john.example.org",
    }
}
`

const noteJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Note",
    "actor": "https://sally.example.org",
    "object": {
        "id": "http://notes.example.com/1",
    }
}
`

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
				t.FailNow()
			}

		})
	}
}

func TestInboxHandler(t *testing.T) {
	t.Parallel()

	i := NewInbox([]string{}, "https", "example.com")

	testResp(t, i, []respTest{
		{followJSON, http.StatusOK, "success_follow_json"},
		{emptyIDFollowJSON, http.StatusOK, "success_follow_json_empty_id"},
		{nullIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_null_id"},
		{missingIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_missing_id"},
		{noteJSON, http.StatusUnsupportedMediaType, "failure_note_json"},
	})

}
