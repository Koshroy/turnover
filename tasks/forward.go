package tasks

import "net/url"

// Forward is a task which forwards a message
type Forward struct {
	TaskID   TaskID
	Activity []byte
	Target   url.URL
}

// ID returns the ID of the Forward task
func (f *Forward) ID() TaskID {
	return f.TaskID
}

// Run forwards the Activity to the Target
func (f *Forward) Run() {

}
