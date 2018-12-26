package tasks

import (
	"bytes"
	"net/http"
	"net/url"
)

// Forward is a task which forwards a message
type Forward struct {
	TaskID   TaskID
	Activity []byte
	Target   url.URL
	client   *http.Client
}

// ID returns the ID of the Forward task
func (f *Forward) ID() TaskID {
	return f.TaskID
}

// Run forwards the Activity to the Target
func (f *Forward) Run() bool {
	reader := bytes.NewReader(f.Activity)
	resp, err := f.client.Post(f.Target.String(), "application/json+ld", reader)
	if err != nil {
		return false
	}
	if resp.StatusCode > 399 {
		return false
	}
	return true
}
