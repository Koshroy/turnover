package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/koshroy/turnover/keystore"
)

type stringTest struct {
	Key, Expected string
}

func testStrings(t *testing.T, data map[string]interface{}, tests []stringTest) {
	for _, tt := range tests {
		t.Run("key_"+tt.Key, func(t *testing.T) {
			val, ok := data[tt.Key]
			if !ok {
				t.Errorf("could not find key %s in %v", tt.Key, data)
				t.FailNow()
			}

			valStr, ok := val.(string)
			if !ok {
				t.Errorf("key %s with value %v is not a string", tt.Key, val)
				t.FailNow()
			}

			if valStr != tt.Expected {
				t.Errorf("for key: %s expected: %s got: %s", tt.Key, tt.Expected, valStr)
			}
		})
	}

}

func TestActorHandler(t *testing.T) {
	t.Parallel()

	a := NewActor("https", "www.example.com", &keystore.Store{})

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

	testStrings(t, respData,
		[]stringTest{
			stringTest{"id", "https://www.example.com/actor"},
			stringTest{"type", "Application"},
			stringTest{"followers", "https://www.example.com/followers"},
			stringTest{"following", "https://www.example.com/following"},
			stringTest{"url", "https://www.example.com/actor"},
			stringTest{"inbox", "https://www.example.com/inbox"},
			stringTest{"outbox", "https://www.example.com/outbox"},
		},
	)
}
