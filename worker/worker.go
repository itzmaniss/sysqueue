package worker

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
	"github.com/itzmaniss/sysqueue/job"
)

type WorkerState string

const (
	WorkerIdle    WorkerState = "Idle"
	WorkerBusy    WorkerState = "Busy"
	WorkerStopped WorkerState = "Stopped"
)

type Worker struct {
	ID       string
	State    WorkerState
	JobQueue chan job.Job
	Jobs     map[string]job.Job
	Lock     *sync.Mutex
}

func NewWorker(jobQueue chan job.Job, jobs map[string]job.Job, lock *sync.Mutex) Worker {
	return Worker{
		ID:       uuid.NewString(),
		State:    WorkerIdle,
		JobQueue: jobQueue,
		Jobs:     jobs,
		Lock:     lock,
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, idx int) {
	defer func() {
		fmt.Printf("\nstopping worker %d", idx)
		wg.Done()
	}()
	for {
		select {
		case j := <-w.JobQueue:
			parts := strings.Fields(j.Command)
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			j.State = job.StateRunning
			j.StartedAt = time.Now()
			w.updateJob(j)
			w.State = WorkerBusy

			err := cmd.Start()
			if err == nil {
				pid := cmd.Process.Pid
				j.PID = &pid
				w.updateJob(j)
				err = cmd.Wait()
			}

			if err != nil {
				log.Println(err)
				j.State = job.StateFailed
			} else {
				j.State = job.StateDone
			}
			j.Duration = time.Since(j.StartedAt)
			w.updateJob(j)
			w.State = WorkerIdle

		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) updateJob(j job.Job) {
	w.Lock.Lock()
	w.Jobs[j.ID] = j
	w.Lock.Unlock()
}
