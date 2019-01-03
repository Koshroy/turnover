package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Koshroy/turnover/subscriptions"
	"github.com/Koshroy/turnover/tasks"
	"github.com/gofrs/uuid"
)

const followJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": "https://activities.example.org/1",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.org/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

var followJSONTarget = url.URL{
	Scheme: "https",
	Host:   "www.example.org",
	Path:   "/inbox",
}

const unFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Undo",
    "id": "https://activities.example.org/1",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Unfollow request",
        "type": "Inbox",
        "id": "https://www.activitypub.org/inbox"
    }
}
`
const emptyIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": "",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.org/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const nullIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "id": null,
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.org/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const missingIDFollowJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Follow",
    "actor": "https://sally.example.org",
    "object": {
        "summary": "Follow request",
        "type": "Inbox",
        "id": "https://www.example.org/inbox",
        "attributedTo": "https://john.example.org"
    }
}
`

const noteJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Note",
    "id": "https://activities.example.org/2",
    "actor": "https://sally.example.org",
    "object": {
        "id": "https://notes.example.org/1"
    }
}
`

const createNoteJSON = `{
    "@context": "https://www.w3.org/ns/activitystreams",
    "@type": "Create",
    "id": "https://activities.example.org/3",
    "actor": "https://sally.otherexample.org",
    "object": {
        "summary": "Note",
        "type": "Note",
        "id": "https://sally.otherexample.org/note/1",
        "attributedTo": "https://john.otherexample.org"
    }
}
`

type mockTransport struct {
	Fallback http.RoundTripper
}

// RoundTrip returns a response in the mock transport for a given request
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "www.w3.org" && req.URL.Path != "/ns/activitystreams" {
		return m.Fallback.RoundTrip(req)
	}

	f, err := os.Open("./testdata/activitystreams.jsonld")
	if err != nil {
		return nil, fmt.Errorf("error opening testdata for mock transport: %v", err)
	}

	s, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting filesize for mock transport: %v", err)
	}

	header := make(http.Header)
	header.Add("Content-Length", fmt.Sprintf("%d", s.Size()))
	header.Add("Content-Type", "application/ld+json")
	header.Add("Date", s.ModTime().Format(time.RFC1123))

	return &http.Response{
		Status:        http.StatusText(http.StatusOK),
		StatusCode:    http.StatusOK,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		ContentLength: s.Size(),
		Request:       req,
		Header:        header,
		Body:          f,
	}, nil
}

type mockQueuer struct {
	enqueued map[uuid.UUID]bool
	finished map[uuid.UUID]bool
}

func newMockQueuer() *mockQueuer {
	return &mockQueuer{
		enqueued: make(map[uuid.UUID]bool),
		finished: make(map[uuid.UUID]bool),
	}
}

func (q *mockQueuer) Enqueue(taskID uuid.UUID) bool {
	q.enqueued[taskID] = true
	return true
}

func (q *mockQueuer) Working() uuid.UUID {
	for tID := range q.enqueued {
		return tID
	}

	return uuid.UUID{}
}

func (q *mockQueuer) ListWorking() []uuid.UUID {
	retTIDs := make([]uuid.UUID, 0)
	for tID := range q.enqueued {
		retTIDs = append(retTIDs, tID)
	}
	return retTIDs
}

func (q *mockQueuer) Finish(taskID uuid.UUID) bool {
	q.finished[taskID] = true
	return true
}

func (q *mockQueuer) ListFinished() []uuid.UUID {
	retTIDs := make([]uuid.UUID, 0)
	for tID := range q.finished {
		retTIDs = append(retTIDs, tID)
	}
	return retTIDs
}

func (q *mockQueuer) ListEnqueues() []uuid.UUID {
	retTIDs := make([]uuid.UUID, 0)
	for tID := range q.enqueued {
		retTIDs = append(retTIDs, tID)
	}
	return retTIDs
}

func (q *mockQueuer) Reset() {
	q.enqueued = make(map[uuid.UUID]bool)
	q.finished = make(map[uuid.UUID]bool)
}

type mockStorer struct {
	storage            *tasks.MemoryStorage
	getCalls, putCalls map[uuid.UUID]bool
}

func newMockStorer() *mockStorer {
	return &mockStorer{
		storage:  tasks.NewMemoryStorage(),
		getCalls: make(map[uuid.UUID]bool),
		putCalls: make(map[uuid.UUID]bool),
	}
}

func (s *mockStorer) Get(taskID uuid.UUID) (tasks.Task, bool) {
	s.getCalls[taskID] = true
	return s.storage.Get(taskID)
}

func (s *mockStorer) Put(task tasks.Task, taskID uuid.UUID) bool {
	s.putCalls[taskID] = true
	return s.storage.Put(task, taskID)
}

func (s *mockStorer) Reset() {
	s.storage = tasks.NewMemoryStorage()
	s.getCalls = make(map[uuid.UUID]bool)
	s.putCalls = make(map[uuid.UUID]bool)
}

func (s *mockStorer) HasPutCall(taskID uuid.UUID) bool {
	_, ok := s.putCalls[taskID]
	return ok
}

type mockForwardManager struct {
	manager *subscriptions.MemManager
	added   map[url.URL]bool
	removed map[url.URL]bool
}

func newMockForwardManager() *mockForwardManager {
	return &mockForwardManager{
		manager: subscriptions.NewMemManager(),
		added:   make(map[url.URL]bool),
		removed: make(map[url.URL]bool),
	}
}

func (fm *mockForwardManager) Reset(initial []url.URL) error {
	manager := subscriptions.NewMemManager()
	for _, target := range initial {
		success := manager.Add(target)
		if !success {
			return fmt.Errorf("could not add %s to forward manager", target.String())
		}
	}
	fm.manager = manager
	fm.added = make(map[url.URL]bool)
	fm.removed = make(map[url.URL]bool)
	return nil
}

func (fm *mockForwardManager) Add(target url.URL) bool {
	success := fm.manager.Add(target)
	if success {
		fm.added[target] = true
	}
	return success
}

func (fm *mockForwardManager) Remove(target url.URL) bool {
	success := fm.manager.Remove(target)
	if success {
		fm.removed[target] = true
	}
	return success
}

func (fm *mockForwardManager) List() []url.URL {
	return fm.manager.List()
}

func (fm *mockForwardManager) IsAdd(url url.URL) bool {
	_, ok := fm.added[url]
	return ok
}

func (fm *mockForwardManager) IsRemove(url url.URL) bool {
	_, ok := fm.removed[url]
	return ok
}

type respTest struct {
	JSONInput  string
	StatusCode int
	TestName   string
}

func testResp(t *testing.T, inbox *Inbox, tests []respTest) {
	for _, tt := range tests {
		t.Run("request_status_code_"+tt.TestName, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.JSONInput))
			w := httptest.NewRecorder()
			inbox.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.StatusCode {
				t.Errorf("expected %d got %d", tt.StatusCode, resp.StatusCode)
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					t.Logf("response body: %v\n", string(respBytes))
				}
				t.FailNow()
			}
		})
	}
}

type forwardTest struct {
	JSONInput      string
	NumEnqueues    int
	ForwardTargets []url.URL
	Adds, Removes  []url.URL
	TestName       string
}

func testForward(t *testing.T, inbox *Inbox, queuer *mockQueuer, storer *mockStorer, forwarder *mockForwardManager, tests []forwardTest) {
	for _, tt := range tests {
		t.Run("request_forward_tasks_"+tt.TestName, func(t *testing.T) {
			queuer.Reset()
			storer.Reset()
			err := forwarder.Reset(tt.ForwardTargets)
			if err != nil {
				t.Errorf("could not seed mock forward manager with targets: %v", err)
				t.FailNow()
			}

			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.JSONInput))
			w := httptest.NewRecorder()
			inbox.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected %d got %d", http.StatusOK, resp.StatusCode)
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					t.Logf("response body: %v\n", string(respBytes))
				}
				t.FailNow()
			}

			enqueues := queuer.ListEnqueues()
			if len(enqueues) != tt.NumEnqueues {
				t.Errorf("expected %d enqueues but got %d enqueues", tt.NumEnqueues, len(enqueues))
				t.FailNow()
			}

			for _, tID := range enqueues {
				if !storer.HasPutCall(tID) {
					t.Errorf("expected %s to be stored but it was not", tID)
					t.FailNow()
				}
			}

			for _, add := range tt.Adds {
				if !forwarder.IsAdd(add) {
					t.Errorf("expected %s to be added but it was not", add.String())
					t.FailNow()
				}
			}

			for _, remove := range tt.Removes {
				if !forwarder.IsRemove(remove) {
					t.Errorf("expected %s to be removed but it was not", remove.String())
					t.FailNow()
				}
			}

		})
	}
}

func TestInboxHandlerResponse(t *testing.T) {
	t.Parallel()

	mockClient := &http.Client{
		Transport: &mockTransport{Fallback: http.DefaultTransport},
	}
	q := newMockQueuer()
	s := newMockStorer()
	fm := newMockForwardManager()
	i := NewInbox([]string{}, "https", "www.example.org", mockClient, q, s, fm)

	baseForwards := []url.URL{{Scheme: "https", Host: "sally.example.org", Path: "/inbox"}}

	testResp(t, i, []respTest{
		{followJSON, http.StatusOK, "success_follow_json"},
		{emptyIDFollowJSON, http.StatusOK, "success_follow_json_empty_id"},
		{nullIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_null_id"},
		{missingIDFollowJSON, http.StatusUnsupportedMediaType, "failure_follow_json_missing_id"},
		{noteJSON, http.StatusUnsupportedMediaType, "failure_note_json"},
		{createNoteJSON, http.StatusOK, "success_create_note_json"},
	})

	testForward(t, i, q, s, fm, []forwardTest{
		{createNoteJSON, 1 * len(baseForwards), baseForwards, []url.URL{}, []url.URL{}, "create_note_json_task_enqueue"},
		{followJSON, 0, []url.URL{}, []url.URL{followJSONTarget}, []url.URL{}, "follow_json_add_forward"},
		{unFollowJSON, 0, baseForwards, []url.URL{}, baseForwards, "unfollow_json_remove_forward"},
	})
}
