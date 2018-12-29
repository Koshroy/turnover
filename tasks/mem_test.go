package tasks

import (
	"testing"

	"github.com/gofrs/uuid"
)

type mockTask struct {
	TaskID uuid.UUID
}

func (t *mockTask) ID() uuid.UUID {
	return t.TaskID
}

func (t *mockTask) Run() error {
	return nil
}

func TestEnqueueWorkingMemQueue(t *testing.T) {
	t.Parallel()

	tID, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	queue := NewMemoryQueue()
	queue.Enqueue(tID)
	workingTID := queue.Working()
	if !uuidEqual(tID, workingTID) {
		t.Errorf("expected Task ID %s found %s", tID, workingTID)
		t.FailNow()
	}
}

func TestEnqueueWorkingListMemQueue(t *testing.T) {
	t.Parallel()

	tID1, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	tID2, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	queue := NewMemoryQueue()
	queue.Enqueue(tID1)
	_ = queue.Working()
	queue.Enqueue(tID2)
	_ = queue.Working()

	tIDs := queue.ListWorking()
	if len(tIDs) != 2 {
		t.Errorf("expected 2 working tasks, got: %d", len(tIDs))
		t.FailNow()
	}
}

func TestFinishMemQueue(t *testing.T) {
	t.Parallel()

	tID1, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	tID2, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	queue := NewMemoryQueue()
	queue.Enqueue(tID1)
	_ = queue.Working()
	queue.Enqueue(tID2)
	workingTID := queue.Working()
	if !uuidEqual(workingTID, tID2) {
		t.Errorf("expected to get working task %s, got: %s", tID2, workingTID)
		t.FailNow()
	}

	success := queue.Finish(workingTID)
	if !success {
		t.Errorf("could not successfully finish %s", workingTID)
		t.FailNow()
	}
}

func TestFinishListMemQueue(t *testing.T) {
	t.Parallel()

	tID1, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	tID2, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	queue := NewMemoryQueue()
	queue.Enqueue(tID1)
	workingTID := queue.Working()
	queue.Enqueue(tID2)
	_ = queue.Working()
	if !uuidEqual(tID1, workingTID) {
		t.Errorf("expected to get working task %s, got: %s", tID1, workingTID)
		t.FailNow()
	}

	success := queue.Finish(workingTID)
	if !success {
		t.Errorf("could not successfully finish %s", workingTID)
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

	tID, err := uuid.NewV4()
	if err != nil {
		t.Errorf("error generating task id: %v", err)
		t.FailNow()
	}

	task := &mockTask{TaskID: tID}
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
	if !uuidEqual(newTask.ID(), tID) {
		t.Errorf("expected to get task of ID %s got ID %s instead", tID, newTask.ID())
		t.FailNow()
	}
}
