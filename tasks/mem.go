package tasks

import "sync"

// MemoryQueue represents a task queue in memory
type MemoryQueue struct {
	waiting chan TaskID

	finishedLock sync.RWMutex
	finished     map[TaskID]bool

	progressLock sync.RWMutex
	progress     map[TaskID]bool
}

// NewMemoryQueue returns a new memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		waiting:  make(chan TaskID, 1),
		finished: make(map[TaskID]bool),
		progress: make(map[TaskID]bool),
	}
}

// Enqueue enques a task
func (m *MemoryQueue) Enqueue(taskID TaskID) bool {
	m.waiting <- taskID
	return true
}

// Working returns a TaskID from the list of waiting tasks and sets
// it into the working state
func (m *MemoryQueue) Working() TaskID {
	m.progressLock.Lock()
	defer m.progressLock.Unlock()

	tID := <-m.waiting
	m.progress[tID] = true
	return tID
}

// ListWorking returns a slice of all TaskIDs in the working state
func (m *MemoryQueue) ListWorking() []TaskID {
	m.progressLock.RLock()
	defer m.progressLock.RUnlock()

	tasks := make([]TaskID, 0)
	for tID := range m.progress {
		tasks = append(tasks, tID)
	}
	return tasks
}

// Finish marks a taskID as finished if it is in progress already
func (m *MemoryQueue) Finish(taskID TaskID) bool {
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

// ListFinished returns a slice of all TaskIDs in the finished state
func (m *MemoryQueue) ListFinished() []TaskID {
	m.finishedLock.RLock()
	defer m.finishedLock.RUnlock()

	tasks := make([]TaskID, 0)
	for tID := range m.finished {
		tasks = append(tasks, tID)
	}
	return tasks
}

// MemoryStorage is an in-memory task storer
type MemoryStorage struct {
	taskStorage map[TaskID]Task
	sync.RWMutex
}

// NewMemoryStorage returns a new MemoryStorage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		taskStorage: make(map[TaskID]Task),
	}
}

// Get returns a task with a given TaskID
func (s *MemoryStorage) Get(taskID TaskID) (Task, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.taskStorage[taskID]
	return t, ok
}

// Put puts a task with the given taskID
func (s *MemoryStorage) Put(task Task, taskID TaskID) bool {
	s.Lock()
	defer s.Unlock()

	s.taskStorage[taskID] = task
	return true
}
