package tasks

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
)

// Forward is a task which forwards a message
type Forward struct {
	TaskID   uuid.UUID
	Activity []byte
	Target   url.URL
	Client   *http.Client
}

// ID returns the ID of the Forward task
func (f *Forward) ID() uuid.UUID {
	return f.TaskID
}

// Run forwards the Activity to the Target
func (f *Forward) Run() error {
	reader := bytes.NewReader(f.Activity)
	resp, err := f.Client.Post(f.Target.String(), "application/ld+json", reader)
	if err != nil {
		return err
	}
	if resp.StatusCode > 399 {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
