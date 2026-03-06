package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkerState string

const (
	WorkerIdle    WorkerState = "Idle"
	WorkerBusy    WorkerState = "Busy"
	WorkerStopped WorkerState = "Stopped"
)

type Worker struct {
	ID    string
	State WorkerState
	Queue *Queue
}

func NewWorker(q *Queue) Worker {
	return Worker{
		ID:    uuid.NewString(),
		State: WorkerIdle,
		Queue: q,
	}
}

func (w *Worker) Start() {
	for {
		job := <-w.Queue.JobQueue
		command := strings.Fields(job.Command)
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		job.State = StateRunning
		job.StartedAt = time.Now()
		w.Queue.Lock.Lock()
		w.Queue.Jobs[job.ID] = job
		w.Queue.Lock.Unlock()
		w.State = WorkerBusy
		err := cmd.Run()

		if err != nil {
			log.Println(err)
			job.State = StateFailed

		} else {
			job.State = StateDone

		}
		job.Duration = time.Since(job.StartedAt)
		w.Queue.Lock.Lock()
		w.Queue.Jobs[job.ID] = job
		w.Queue.Lock.Unlock()
		w.State = WorkerIdle
	}
}
