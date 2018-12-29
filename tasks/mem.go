package tasks

import (
	"github.com/gofrs/uuid"

	"sync"
)

// MemoryQueue represents a task queue in memory
type MemoryQueue struct {
	waiting chan uuid.UUID

	finishedLock sync.RWMutex
	finished     map[uuid.UUID]bool

	progressLock sync.RWMutex
	progress     map[uuid.UUID]bool
}

// NewMemoryQueue returns a new memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		waiting:  make(chan uuid.UUID, 1),
		finished: make(map[uuid.UUID]bool),
		progress: make(map[uuid.UUID]bool),
	}
}

// Enqueue enques a task
func (m *MemoryQueue) Enqueue(taskID uuid.UUID) bool {
	m.waiting <- taskID
	return true
}

// Working returns a uuid.UUID from the list of waiting tasks and sets
// it into the working state
func (m *MemoryQueue) Working() uuid.UUID {
	m.progressLock.Lock()
	defer m.progressLock.Unlock()

	tID := <-m.waiting
	m.progress[tID] = true
	return tID
}

// ListWorking returns a slice of all uuid.UUIDs in the working state
func (m *MemoryQueue) ListWorking() []uuid.UUID {
	m.progressLock.RLock()
	defer m.progressLock.RUnlock()

	tasks := make([]uuid.UUID, 0)
	for tID := range m.progress {
		tasks = append(tasks, tID)
	}
	return tasks
}

// Finish marks a taskID as finished if it is in progress already
func (m *MemoryQueue) Finish(taskID uuid.UUID) bool {
	m.progressLock.RLock()
	if _, ok := m.progress[taskID]; !ok {
		m.progressLock.RUnlock()
		return false
	}
	m.progressLock.RUnlock()

	m.finishedLock.Lock()
	defer m.finishedLock.Unlock()

	delete(m.progress, taskID)
	m.finished[taskID] = true
	return true
}

// ListFinished returns a slice of all uuid.UUIDs in the finished state
func (m *MemoryQueue) ListFinished() []uuid.UUID {
	m.finishedLock.RLock()
	defer m.finishedLock.RUnlock()

	tasks := make([]uuid.UUID, 0)
	for tID := range m.finished {
		tasks = append(tasks, tID)
	}
	return tasks
}

// MemoryStorage is an in-memory task storer
type MemoryStorage struct {
	taskStorage map[uuid.UUID]Task
	sync.RWMutex
}

// NewMemoryStorage returns a new MemoryStorage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		taskStorage: make(map[uuid.UUID]Task),
	}
}

// Get returns a task with a given uuid.UUID
func (s *MemoryStorage) Get(taskID uuid.UUID) (Task, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.taskStorage[taskID]
	return t, ok
}

// Put puts a task with the given taskID
func (s *MemoryStorage) Put(task Task, taskID uuid.UUID) bool {
	s.Lock()
	defer s.Unlock()

	s.taskStorage[taskID] = task
	return true
}
