package tasks

import "testing"

type mockTask struct {
	TaskID TaskID
}

func (t *mockTask) ID() TaskID {
	return t.TaskID
}

func (t *mockTask) Run() error {
	return nil
}

func TestEnqueueWorkingMemQueue(t *testing.T) {
	t.Parallel()
	queue := NewMemoryQueue()
	queue.Enqueue("a")
	tID := queue.Working()
	if tID != "a" {
		t.Errorf("expected Task ID a found %s", tID)
		t.FailNow()
	}
}

func TestEnqueueWorkingListMemQueue(t *testing.T) {
	t.Parallel()
	queue := NewMemoryQueue()
	queue.Enqueue("a")
	_ = queue.Working()
	queue.Enqueue("b")
	_ = queue.Working()

	tIDs := queue.ListWorking()
	if len(tIDs) != 2 {
		t.Errorf("expected 2 working tasks, got: %d", len(tIDs))
		t.FailNow()
	}
}

func TestFinishMemQueue(t *testing.T) {
	t.Parallel()
	queue := NewMemoryQueue()
	queue.Enqueue("a")
	_ = queue.Working()
	queue.Enqueue("b")
	tID := queue.Working()
	if tID != "b" {
		t.Errorf("expected to get working task a, got: %s", tID)
		t.FailNow()
	}

	success := queue.Finish(tID)
	if !success {
		t.Errorf("could not successfully finish %s", tID)
		t.FailNow()
	}
}

func TestFinishListMemQueue(t *testing.T) {
	t.Parallel()
	queue := NewMemoryQueue()
	queue.Enqueue("a")
	tID := queue.Working()
	queue.Enqueue("b")
	_ = queue.Working()
	if tID != "a" {
		t.Errorf("expected to get working task a, got: %s", tID)
		t.FailNow()
	}

	success := queue.Finish(tID)
	if !success {
		t.Errorf("could not successfully finish %s", tID)
		t.FailNow()
	}

	finishList := queue.ListFinished()
	if len(finishList) != 1 {
		t.Errorf("expected 1 task to be finished, actually %d tasks were finished", len(finishList))
		t.FailNow()
	}
}

func TestMemStorage(t *testing.T) {
	t.Parallel()
	task := &mockTask{TaskID: "a"}
	store := NewMemoryStorage()

	res := store.Put(task, task.ID())
	if !res {
		t.Error("error putting task in mem store")
		t.FailNow()
	}

	newTask, ok := store.Get(task.ID())
	if !ok {
		t.Errorf("could not get task with ID %s in store", task.ID())
		t.FailNow()
	}
	if newTask.ID() != "a" {
		t.Errorf("expected to get task of ID a got ID %s instead", newTask.ID())
		t.FailNow()
	}
}
