package controllers

import (
	"encoding/json"
	"log"
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
		"type":      "Application",
		"following": a.routeURL("/following", "").String(),
		"followers": a.routeURL("/followers", "").String(),
		"inbox":     a.routeURL("/inbox", "").String(),
		"outbox":    a.routeURL("/outbox", "").String(),
		"id":        a.routeURL("/actor", "").String(),
		"name":      "turnover relay",
		"summary":   "An ActivityPub Relay",
		"url":       a.routeURL("/actor", "").String(),
		"publicKey": map[string]string{
			"publicKeyPem": "",
			"owner":        a.routeURL("/actor", "").String(),
			"id":           a.routeURL("/actor", "#main-key").String(),
		},
	}

	b, err := json.Marshal(actorData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Printf("error writing response: %v\n", err)
	}
}

func (a Actor) routeURL(path, fragment string) *url.URL {
	return &url.URL{
		Scheme:   a.Scheme,
		Host:     a.Domain,
		Path:     path,
		Fragment: fragment,
	}
}
