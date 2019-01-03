package models

// Activity represents an ActivityPub Activity
type Activity struct {
	// Object, Actor, and Type are required fields
	Object []Activity `json:"https://www.w3.org/ns/activitystreams#object,omitempty"`
	Actor  []Activity `json:"https://www.w3.org/ns/activitystreams#actor,omitempty"`
	Type   []string   `json:"@type,omitempty"`
	// We need to be able to distinguish between omitted IDs and blank IDs
	ID *string `json:"@id,omitempty"`

	// We don't care about the following fields so they are omittable and
	// we simply pass them on
	Target     interface{} `json:"https://www.w3.org/ns/activitystreams#target,omitempty"`
	Result     interface{} `json:"https://www.w3.org/ns/activitystreams#result,omitempty"`
	Origin     interface{} `json:"https://www.w3.org/ns/activitystreams#origin,omitempty"`
	Instrument interface{} `json:"https://www.w3.org/ns/activitystreams#instrument,omitempty"`
}
