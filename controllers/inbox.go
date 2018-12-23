package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/Koshroy/turnover/models"
	"github.com/piprate/json-gold/ld"
)

const maxActivitySz = 16 * (1 << 20) // 16 MB
const followIRI = "https://www.w3.org/ns/activitystreams#Follow"
const unfollowIRI = "https://www.w3.org/ns/activitystreams#Unfollow"
const createIRI = "https://www.w3.org/ns/activitystreams#Create"
const readIRI = "https://www.w3.org/ns/activitystreams#Read"
const updateIRI = "https://www.w3.org/ns/activitystreams#Update"
const deleteIRI = "https://www.w3.org/ns/activitystreams#Delete"

// ErrUnsupportedActivityType is returned when the activity
// contains a type that is not Follow, Create, Read, Update, Delete, or Unfollow
// or is a multi-type activity
var ErrUnsupportedActivityType = errors.New("unsupported activity type")

// ErrNullIDUnsupported is returned when the ID is specifically missing or set to null
var ErrNullIDUnsupported = errors.New("activity id cannot be null or missing")

// ErrIncorrectFollow is returned when a non-inbox endpoint is attempted to be followed
var ErrIncorrectFollow = errors.New("cannot follow this resource")

// Inbox is a controller that controls the Inbox endpoint
type Inbox struct {
	whitelist      []string
	loader         *ld.RFC7324CachingDocumentLoader
	proc           *ld.JsonLdProcessor
	opts           *ld.JsonLdOptions
	scheme, domain string
}

// NewInbox creates a new Inbox controller
func NewInbox(whitelist []string, scheme, domain string) *Inbox {
	loader := ld.NewRFC7324CachingDocumentLoader(nil)
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = loader

	return &Inbox{
		whitelist: whitelist,
		loader:    loader,
		proc:      ld.NewJsonLdProcessor(),
		opts:      opts,
		scheme:    scheme,
		domain:    domain,
	}
}

func (i Inbox) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxActivitySz)
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var raw map[string]interface{}
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	expanded, err := i.proc.Expand(raw, i.opts)
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	for _, rawActivity := range expanded {
		activity, typeOk := rawActivity.(map[string]interface{})
		if !typeOk {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				log.Printf("error writing response: %v\n", err)
			}
			return
		}
		hydrated, err := hydrateActivity(activity)
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				log.Printf("error writing response: %v\n", err)
			}
			return
		}

		myInboxURI := i.routeURL("/inbox", "").String()
		for _, hydratedType := range hydrated.Type {
			if hydratedType == followIRI || hydratedType == unfollowIRI {
				for _, objectActivity := range hydrated.Object {
					if objectActivity.ID == nil ||
						(*objectActivity.ID != myInboxURI) {
						w.WriteHeader(http.StatusMethodNotAllowed)
						_, err := w.Write([]byte("follows and unfollows can only be to " + myInboxURI))
						if err != nil {
							log.Printf("error writing response: %v\n", err)
						}
						return
					}
				}
			}
		}
	}
}

func hydrateActivity(raw map[string]interface{}) (*models.Activity, error) {
	// This function is kinda jank because it marshals a raw interface
	// then unmarshals it into a models.Activity type. There is probably
	// a better way to do this, but rather than write complicated code to do that
	// for now we'll do this. If this ends up being a speed bottleneck, we can
	// write the bespoke logic for it

	activityBytes, err := json.Marshal(raw)
	if err != nil {
		// TODO: let's wrap this error to make this a better function
		return nil, err
	}

	var activity models.Activity
	err = json.Unmarshal(activityBytes, &activity)
	if err != nil {
		return nil, ErrUnsupportedActivityType
	}

	for _, activityType := range activity.Type {
		if activityType != followIRI &&
			activityType != createIRI &&
			activityType != updateIRI &&
			activityType != readIRI &&
			activityType != deleteIRI &&
			activityType != unfollowIRI {
			return nil, ErrUnsupportedActivityType
		}
	}

	// We disallow null IDs
	if activity.ID == nil {
		return nil, ErrNullIDUnsupported
	}

	return &activity, nil
}

func (i Inbox) routeURL(path, fragment string) *url.URL {
	return &url.URL{
		Scheme:   i.scheme,
		Host:     i.domain,
		Path:     path,
		Fragment: fragment,
	}
}
