package middleware

import (
	"net/http"
	"strings"
)

// ActivityPubHeaders is a middleware which fails the request if ActivityPub headers are
// not present in the request
func ActivityPubHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerOk := false

		acceptTypes, ok := r.Header["Accept"]
		if ok {
			for _, acceptType := range acceptTypes {
				if strings.Contains(acceptType, "application/ld+json") {
					headerOk = true
				}
			}
		}

		if !headerOk {
			contentTypes, ok := r.Header["Content-Type"]
			if ok {
				for _, contentType := range contentTypes {
					if strings.Contains(contentType, "application/activity+json") {
						headerOk = true
					}
				}
			}
		}

		if !headerOk {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}
