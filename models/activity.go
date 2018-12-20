package models

// Activity represents an ActivityPub Activity
type Activity struct {
	// Object and Type are required fields
	Object, Type string
	// We need to be able to distinguish between omitted IDs and blank IDs
	ID *string `json:"id,omitempty"`

	// We don't care about the following fields so they are omittable and
	// we simply pass them on
	Actor      interface{} `json:"actor,omitempty"`
	Target     interface{} `json:"target,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	Origin     interface{} `json:"origin,omitempty"`
	Instrument interface{} `json:"instrument,omitempty"`
}
