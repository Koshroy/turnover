package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Koshroy/turnover/keystore"
)

type stringTest struct {
	Key, Expected string
}

func testStrings(t *testing.T, data map[string]interface{}, tests []stringTest) {
	for _, tt := range tests {
		t.Run("response_key_"+tt.Key, func(t *testing.T) {
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

func checkPubKeyPem(t *testing.T, resp map[string]interface{}, pubKeyPemStr string) {
	t.Run("response_privkey_pem", func(t *testing.T) {
		pubKeyBlockRaw, ok := resp["publicKey"]
		if !ok {
			t.Errorf("key publicKey not found in response")
			t.FailNow()
		}

		pubKeyBlock := pubKeyBlockRaw.(map[string]interface{})
		pubKeyPemRaw := pubKeyBlock["publicKeyPem"]
		pubKeyPem := pubKeyPemRaw.(string)
		if pubKeyPem != pubKeyPemStr {
			t.Errorf("public key PEM incorrect expected: %s got: %s", pubKeyPemStr, pubKeyPem)
		}
	})
}

func TestActorHandler(t *testing.T) {
	t.Parallel()

	a := NewActor("https", "www.example.com", keystore.MockStore())

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
			{"id", "https://www.example.com/actor"},
			{"type", "Application"},
			{"followers", "https://www.example.com/followers"},
			{"following", "https://www.example.com/following"},
			{"url", "https://www.example.com/actor"},
			{"inbox", "https://www.example.com/inbox"},
			{"outbox", "https://www.example.com/outbox"},
		},
	)

	checkPubKeyPem(t, respData, string(keystore.MockPubKey))
}
