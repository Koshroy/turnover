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
        "type": "Note",
        "id": "http://notes.example.com/1",
        "attributedTo": "http://john.example.org",
        "content": "My note"
    }    
}
`

func TestInboxHandler(t *testing.T) {
	// t.Parallel()

	i := NewInbox([]string{})

	req := httptest.NewRequest("POST", "/", strings.NewReader(followJSON))
	w := httptest.NewRecorder()

	i.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK got %d", resp.StatusCode)
		t.FailNow()
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	t.Errorf("error reading bytes from response %v", err)
	// 	t.FailNow()
	// }

	// var respData map[string]interface{}

	// err = json.Unmarshal(bodyBytes, &respData)
	// if err != nil {
	// 	t.Error("could not unmarshal response data")
	// 	t.FailNow()
	// }
}
