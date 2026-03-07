package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
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

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, idx int) {
	defer func() {
		fmt.Printf("\nstopping worker %d", idx)
		wg.Done()
	}()
	for {
		select {
		case job := <-w.Queue.JobQueue:

			command := strings.Fields(job.Command)
			cmd := exec.Command(command[0], command[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			job.State = StateRunning
			job.StartedAt = time.Now()
			w.UpdateJob(job)
			w.State = WorkerBusy

			err := cmd.Start()
			if err == nil {
				PID := cmd.Process.Pid
				job.PID = &PID
				w.UpdateJob(job)
				err = cmd.Wait()
			}

			if err != nil {
				log.Println(err)
				job.State = StateFailed

			} else {
				job.State = StateDone

			}
			job.Duration = time.Since(job.StartedAt)
			w.UpdateJob(job)
			w.State = WorkerIdle
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) UpdateJob(job Job) {
	w.Queue.Lock.Lock()
	w.Queue.Jobs[job.ID] = job
	w.Queue.Lock.Unlock()
}
