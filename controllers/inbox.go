package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Koshroy/turnover/models"
	"github.com/piprate/json-gold/ld"
)

const maxActivitySz = 16 * (1 << 20) // 16 MB
const objectIRI = "https://www.w3.org/ns/activitystreams#object"
const followIRI = "https://www.w3.org/ns/activitystreams#Follow"
const unfollowIRI = "https://www.w3.org/ns/activitystreams#Unfollow"
const createIRI = "https://www.w3.org/ns/activitystreams#Create"
const readIRI = "https://www.w3.org/ns/activitystreams#Read"
const updateIRI = "https://www.w3.org/ns/activitystreams#Update"
const deleteIRI = "https://www.w3.org/ns/activitystreams#Delete"

// ErrNoObject is returned when an activity has no Object field
var ErrNoObject = errors.New("no object field found in activity")

// ErrNullID is returned when the activity has a null Id field
var ErrNullID = errors.New("activity has null id")

// ErrUnsupportedActivityType is returned when the activity
// contains a type that is not Follow, Create, Read, Update, Delete, or Unfollow
var ErrUnsupportedActivityType = errors.New("unsupported activity type")

// Inbox is a controller that controls the Inbox endpoint
type Inbox struct {
	whitelist []string
	loader    *ld.RFC7324CachingDocumentLoader
	proc      *ld.JsonLdProcessor
	opts      *ld.JsonLdOptions
}

// NewInbox creates a new Inbox controller
func NewInbox(whitelist []string) *Inbox {
	loader := ld.NewRFC7324CachingDocumentLoader(nil)
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = loader

	return &Inbox{
		whitelist: whitelist,
		loader:    loader,
		proc:      ld.NewJsonLdProcessor(),
		opts:      opts,
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
	fmt.Printf("raw: %v\n", raw)

	expanded, err := i.proc.Expand(raw, i.opts)
	if err != nil {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	fmt.Printf("expanded: %v\n", expanded)

	for _, rawActivity := range expanded {
		activity, typeOk := rawActivity.(map[string]interface{})
		if !typeOk {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		_, err = hydrateActivity(activity)
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
	}
}

func hydrateActivity(raw map[string]interface{}) (*models.Activity, error) {
	if _, ok := raw[objectIRI]; !ok {
		return nil, ErrNoObject
	}

	// raw activity must have @type or else it would not have expanded properly
	aType := raw["@type"]
	// if aType != followIRI &&
	// 	aType != createIRI &&
	// 	aType != updateIRI &&
	// 	aType != readIRI &&
	// 	aType != deleteIRI &&
	// 	aType != unfollowIRI {
	// 	return nil, ErrUnsupportedActivityType
	// }

	typeStr := ""
	switch aType {
	case createIRI:
		typeStr = "create"
	case updateIRI:
		typeStr = "follow"
	case readIRI:
		typeStr = "read"
	case deleteIRI:
		typeStr = "delete"
	case unfollowIRI:
		typeStr = "unfollow"
	default:
		return nil, ErrUnsupportedActivityType

	}

	return &models.Activity{
		Type: typeStr,
	}, nil
}
