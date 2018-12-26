package tasks

// TaskID is the ID type of a task
type TaskID string

// Task is an asynch task
type Task interface {
	ID() TaskID
	Run() bool
}

// Queuer can enqueue and dequeue tasks
type Queuer interface {
	Enqueue(taskID TaskID) bool
	Working() TaskID
	ListWorking() []TaskID
	Finish(taskID TaskID) bool
	ListFinished() []TaskID
}

// Storer can load and store task data
type Storer interface {
	Get(taskID TaskID) (Task, bool)
	Put(task Task, taskID TaskID) bool
}
