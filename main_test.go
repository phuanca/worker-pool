package workerpool

import (
	"testing"
	"time"
)

type dummyTask struct {
	done         bool
	taskDuration time.Duration
}

func (d *dummyTask) Execute() {
	time.Sleep(d.taskDuration)
	d.done = true
}

func TestCreateService(t *testing.T) {
	// given
	taskTime := time.Duration(100 * time.Millisecond)
	task := dummyTask{done: false, taskDuration: taskTime}
	workers := 1
	queueDepth := 1

	// when
	undertest := NewService(workers, queueDepth)
	defer undertest.Exit()

	undertest.Run()
	undertest.Send(&task)

	time.Sleep(120 * time.Millisecond) // wait taskTime +20ms

	// then
	if !task.done {
		t.Errorf("task incomplete")
	}
}

func TestDelayedRunDispacher(t *testing.T) {
	// given
	numOfTaks := 4
	taskDuration := time.Duration(100 * time.Millisecond)
	workers := 1
	queueDepth := 0 //unbuffered channel

	// when
	undertest := NewService(workers, queueDepth)
	defer undertest.Exit()

	// run dispacher after 120 ms.
	go func(wp *Service) {
		wait := time.After(120 * time.Millisecond)
		<-wait
		wp.Run()

	}(undertest)

	allTasks := make([]*dummyTask, numOfTaks)

	go func(at []*dummyTask) {
		for i := 0; i < numOfTaks; i++ {
			task := dummyTask{done: false, taskDuration: taskDuration}
			at[i] = &task
			undertest.Send(&task)
		}
	}(allTasks)

	// after 420ms all results must be completed.
	time.Sleep(420 * time.Millisecond)

	doneCount := 0
	for i := range allTasks {
		if allTasks[i].done {
			doneCount++
		}
	}

	if doneCount == numOfTaks {
		t.Errorf("must be at least one task incomplete. Count: %d", doneCount)
	}

}
