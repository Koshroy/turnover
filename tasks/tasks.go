package tasks

import "github.com/gofrs/uuid"

// Task is an asynch task
type Task interface {
	ID() uuid.UUID
	Run() error
}

// Queuer can enqueue and dequeue tasks
type Queuer interface {
	Enqueue(taskID uuid.UUID) bool
	Working() uuid.UUID
	ListWorking() []uuid.UUID
	Finish(taskID uuid.UUID) bool
	ListFinished() []uuid.UUID
}

// Storer can load and store task data
type Storer interface {
	Get(taskID uuid.UUID) (Task, bool)
	Put(task Task, taskID uuid.UUID) bool
}

// NewTaskID creates a new TaskID
func NewTaskID() (uuid.UUID, error) {
	return uuid.NewV4()
}
