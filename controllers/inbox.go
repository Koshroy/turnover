package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/Koshroy/turnover/models"
	"github.com/Koshroy/turnover/subscriptions"
	"github.com/Koshroy/turnover/tasks"
	"github.com/piprate/json-gold/ld"
)

const maxActivitySz = 16 * (1 << 20) // 16 MB
const followIRI = "https://www.w3.org/ns/activitystreams#Follow"
const undoIRI = "https://www.w3.org/ns/activitystreams#Undo"
const createIRI = "https://www.w3.org/ns/activitystreams#Create"
const readIRI = "https://www.w3.org/ns/activitystreams#Read"
const updateIRI = "https://www.w3.org/ns/activitystreams#Update"
const deleteIRI = "https://www.w3.org/ns/activitystreams#Delete"

// ErrUnsupportedActivityType is returned when the activity
// contains a type that is not Follow, Create, Read, Update, Delete, or Undo
// or is a multi-type activity
var ErrUnsupportedActivityType = errors.New("unsupported activity type")

// ErrNullIDUnsupported is returned when the ID is specifically missing or set to null
var ErrNullIDUnsupported = errors.New("activity id cannot be null or missing")

// ErrIncorrectFollow is returned when a non-inbox endpoint is attempted to be followed
var ErrIncorrectFollow = errors.New("cannot follow this resource")

const (
	followDecision = iota
	unfollowDecision
	otherDecision
	invalidDecision
)

// Inbox is a controller that controls the Inbox endpoint
type Inbox struct {
	whitelist      []string
	loader         *ld.RFC7324CachingDocumentLoader
	proc           *ld.JsonLdProcessor
	opts           *ld.JsonLdOptions
	scheme, domain string
	queuer         tasks.Queuer
	storer         tasks.Storer
	manager        subscriptions.Manager
}

// NewInbox creates a new Inbox controller
func NewInbox(
	whitelist []string,
	scheme, domain string,
	client *http.Client,
	queuer tasks.Queuer,
	storer tasks.Storer,
	manager subscriptions.Manager,
) *Inbox {
	loader := ld.NewRFC7324CachingDocumentLoader(client)
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = loader

	return &Inbox{
		whitelist: whitelist,
		loader:    loader,
		proc:      ld.NewJsonLdProcessor(),
		opts:      opts,
		scheme:    scheme,
		domain:    domain,
		queuer:    queuer,
		storer:    storer,
		manager:   manager,
	}
}

func errorResponse(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	w.WriteHeader(statusCode)
	_, writeErr := w.Write([]byte(err.Error()))
	if writeErr != nil {
		log.Printf("error writing response: %v\n", writeErr)
	}
}

func (i Inbox) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxActivitySz)
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		errorResponse(w, r, http.StatusNotAcceptable, errors.New("could not read http response"))
		return
	}

	var raw map[string]interface{}
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		errorResponse(w, r, http.StatusUnsupportedMediaType, errors.New("incorrect json request format"))
		return
	}

	expanded, err := i.proc.Expand(raw, i.opts)
	if err != nil {
		errorResponse(w, r, http.StatusUnsupportedMediaType, err)
		return
	}

	hydratedActivities := make([]*models.Activity, 0)
	for _, rawActivity := range expanded {
		activity, typeOk := rawActivity.(map[string]interface{})
		if !typeOk {
			errorResponse(w, r,
				http.StatusUnsupportedMediaType,
				errors.New("activity could not be converted properly"),
			)
			return
		}
		hydrated, err := hydrateActivity(activity)
		if err != nil {
			errorResponse(w, r,
				http.StatusUnsupportedMediaType,
				errors.New("could not convert json to activitypub form: "+err.Error()),
			)
			return
		}

		hydratedActivities = append(hydratedActivities, hydrated)
	}

	myInboxURI := i.routeURL("/inbox", "")

	for _, activity := range hydratedActivities {
		decision := followOrUnfollowActivity(*activity, myInboxURI.String())
		switch decision {
		case followDecision:
			objIDURLs, err := getObjectIDURLs(*activity)
			if err != nil {
				errorResponse(w, r,
					http.StatusUnsupportedMediaType,
					fmt.Errorf("invalid follow target: %v", err),
				)
				return
			}

			if !filterURL(objIDURLs, myInboxURI) {
				errorResponse(w, r,
					http.StatusUnsupportedMediaType,
					fmt.Errorf("follow targets can only be the inbox of this server"),
				)
				return
			}

			actIDURLs, err := getActorIDURLs(*activity)
			if err != nil {
				errorResponse(w, r,
					http.StatusUnsupportedMediaType,
					fmt.Errorf("invalid follow source: %v", err),
				)
			}

			for _, aIDURL := range actIDURLs {
				ok := i.manager.Add(*aIDURL)
				if !ok {
					errorResponse(w, r,
						http.StatusUnsupportedMediaType,
						fmt.Errorf("could not follow URL: %s", aIDURL.String()),
					)
					return
				}
			}
		case unfollowDecision:
			for _, object := range activity.Object {
				objIDURLs, err := getObjectIDURLs(object)
				if err != nil {
					errorResponse(w, r,
						http.StatusUnsupportedMediaType,
						fmt.Errorf("invalid unfollow target: %v", err),
					)
					return
				}

				if !filterURL(objIDURLs, myInboxURI) {
					errorResponse(w, r,
						http.StatusUnsupportedMediaType,
						fmt.Errorf("unfollow targets can only be the inbox of this server"),
					)
					return
				}
			}
		case otherDecision:
			taskID, err := tasks.NewTaskID()
			if err != nil {
				log.Printf("error generating task ID: %v\n", err)
				errorResponse(w, r, http.StatusInternalServerError, err)
				continue
			}

			activityBytes, err := json.Marshal(activity)
			if err != nil {
				log.Printf("error marshalling activity: %v\n", err)
				errorResponse(w, r, http.StatusInternalServerError, err)
				continue
			}

			for _, target := range i.manager.List() {
				forward := &tasks.Forward{
					Target:   target,
					TaskID:   taskID,
					Activity: activityBytes,
					Client:   http.DefaultClient,
				}

				success := i.storer.Put(forward, taskID)
				if !success {
					log.Println("error storing activity")
					w.WriteHeader(http.StatusInternalServerError)
					_, writeErr := w.Write([]byte("could not store task information"))
					if writeErr != nil {
						log.Printf("error writing response: %v\n", writeErr)
					}
					continue
				}

				success = i.queuer.Enqueue(taskID)
				if !success {
					// TODO: should we delete the task storage if we could not enqueue it properly?
					log.Println("error enqueuing forward activity")
					w.WriteHeader(http.StatusInternalServerError)
					_, writeErr := w.Write([]byte("could not enqueue forward activity"))
					if writeErr != nil {
						log.Printf("error writing response: %v\n", writeErr)
					}
					continue
				}
			}
		case invalidDecision:
			errorResponse(w, r, http.StatusUnsupportedMediaType, errors.New("invalid activitypub type"))
			return
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
		return nil, fmt.Errorf("could not marshall JSON properly: %v", err)
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
			activityType != undoIRI {
			return nil, ErrUnsupportedActivityType
		}
	}

	// We disallow null IDs
	if activity.ID == nil {
		return nil, ErrNullIDUnsupported
	}

	return &activity, nil
}

func followOrUnfollowActivity(activity models.Activity, inboxURL string) int {
	for _, aType := range activity.Type {
		if aType == followIRI {
			for _, object := range activity.Object {
				if object.ID != nil && *object.ID == inboxURL {
					return followDecision
				}
			}
			return invalidDecision
		}

		if aType == undoIRI {
			for _, object := range activity.Object {
				for _, oType := range object.Type {
					if oType == followIRI {
						if object.ID != nil && *object.ID == inboxURL {
							return unfollowDecision
						}

						return invalidDecision
					}
				}
			}
		}
	}
	return otherDecision
}

func getObjectIDURLs(activity models.Activity) ([]*url.URL, error) {
	retURLs := make([]*url.URL, 0)

	for _, object := range activity.Object {
		objID := object.ID
		if objID == nil {
			continue
		}
		objIDURL, err := url.Parse(*objID)
		if err != nil {
			return nil, fmt.Errorf("activity has object with ID %v that is not a valid URL: %v", *objID, err)
		}
		retURLs = append(retURLs, objIDURL)
	}

	if len(retURLs) == 0 {
		return retURLs, errors.New("no objects found in activity")
	}

	return retURLs, nil

}

func getActorIDURLs(activity models.Activity) ([]*url.URL, error) {
	retURLs := make([]*url.URL, 0)

	for _, actor := range activity.Actor {
		if actor.ID == nil {
			continue
		}

		aIDURL, err := url.Parse(*actor.ID)
		if err != nil {
			return nil, fmt.Errorf("activity has type with ID %v that is not a valid URL: %v", *actor.ID, err)
		}
		retURLs = append(retURLs, aIDURL)
	}

	if len(retURLs) == 0 {
		return retURLs, errors.New("no actors found in activity")
	}

	return retURLs, nil
}

func filterURL(urls []*url.URL, needle *url.URL) bool {
	for _, url := range urls {
		if *needle == *url {
			return true
		}
	}

	return false
}

func (i Inbox) routeURL(path, fragment string) *url.URL {
	return &url.URL{
		Scheme:   i.scheme,
		Host:     i.domain,
		Path:     path,
		Fragment: fragment,
	}
}
