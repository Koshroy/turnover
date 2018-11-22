package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/koshroy/turnover/keystore"
)

func TestActorHandler(t *testing.T) {
	a := NewActor("https", "example.com", &keystore.Store{})

	req := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	a.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK got %d", resp.StatusCode)
		t.FailNow()
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("error reading bytes from response %v", err)
		t.FailNow()
	}

	var respData map[string]interface{}

	err = json.Unmarshal(bodyBytes, &respData)
	if err != nil {
		t.Error("could not unmarshal response data")
		t.FailNow()
	}

	followersIRIVal, ok := respData["followers"]
	if !ok {
		t.Errorf("could not find 'followers' key in result JSON: %v", respData)
		t.FailNow()
	}

	followersIRI, ok := followersIRIVal.(string)
	if !ok {
		t.Errorf("'followers' key is not a string: %v", followersIRIVal)
		t.FailNow()
	}

	if followersIRI != "https://example.com/followers" {
		t.Errorf("followersIRI expected: https://example.com/followers actual: %s", followersIRI)
	}
}
