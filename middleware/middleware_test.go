package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func TestActivityPubHeaders(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name             string
		inputContentType []string
		inputAccept      []string
		want             int
	}{
		{
			"should accept requests with activitypub accept headers",
			[]string{"application/json"},
			[]string{"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},
			http.StatusOK,
		},
		{
			"should accept requests with activitypub accept headers with charset in content type",
			[]string{"application/json; charset=utf-8"},
			[]string{"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},
			http.StatusOK,
		},
		{
			"should accept requests with activitypub accept headers but weird content type",
			[]string{"text/html"},
			[]string{"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},
			http.StatusOK,
		},
		{
			"should not accept requests with no accept types and wrong content type",
			[]string{"application/json"},			
			[]string{""},
			http.StatusUnsupportedMediaType,
		},
		{
			"should not accept requests with wrong accept types and wrong content type",
			[]string{"application/json"},			
			[]string{"application/json"},
			http.StatusUnsupportedMediaType,
		},
		{
			"should accept requests with activitypub content type",
			[]string{"application/activity+json"},		
			[]string{""},
			http.StatusOK,
		},
		{
			"should accept requests with activitypub content type and accept header",
			[]string{"application/activity+json"},
			[]string{"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},			
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Use(ActivityPubHeaders)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {})

			req, _ := http.NewRequest("GET", "/", nil)
			req.Header["Content-Type"] = tt.inputContentType
			req.Header["Accept"] = tt.inputAccept

			r.ServeHTTP(recorder, req)
			res := recorder.Result()

			if res.StatusCode != tt.want {
				t.Errorf("response is incorrect, got %d, want %d", recorder.Code, tt.want)
			}
		})
	}
}
