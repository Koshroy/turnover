package subscriptions

import "net/url"

// Manager manages a set of subscriptions for forwarding
type Manager interface {
	Add(target url.URL) bool
	Remove(target url.URL) bool
	List() []url.URL
}
