package queue

import (
	"sync"

	"github.com/itzmaniss/sysqueue/job"
)

type Queue struct {
	JobQueue chan job.Job
	Jobs     map[string]job.Job
	Lock     sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		JobQueue: make(chan job.Job),
		Jobs:     make(map[string]job.Job),
	}
}

func (q *Queue) Enqueue(command string) job.Job {
	j := job.NewJob(command)
	q.Lock.Lock()
	q.Jobs[j.ID] = j
	q.Lock.Unlock()
	q.JobQueue <- j
	return j
}
