package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/koshroy/turnover/keystore"
)

// Actor is the controller logic for the /actor endpoint
type Actor struct {
	Scheme, Domain string
	Store          *keystore.Store
}

// NewActor creates a new Actor
func NewActor(scheme, domain string, store *keystore.Store) Actor {
	return Actor{
		Scheme: scheme,
		Domain: domain,
		Store:  store,
	}
}

func (a Actor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actorData := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/ns/activitystreams",
			"https://web-payments.org/contexts/security-v1.jsonld",
		},
		"type":      "application",
		"followers": a.routeURL("/followers").String(),
	}

	b, err := json.Marshal(actorData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func (a Actor) routeURL(path string) *url.URL {
	return &url.URL{
		Scheme: a.Scheme,
		Host:   a.Domain,
		Path:   path,
	}
}
