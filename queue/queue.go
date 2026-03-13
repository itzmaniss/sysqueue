package queue

import (
	"sync"

	"github.com/itzmaniss/sysqueue/job"
	"github.com/itzmaniss/sysqueue/metrics"
)

type Queue struct {
	JobQueue chan job.Job
	Jobs     map[string]job.Job
	Lock     sync.Mutex
	Metrics  *metrics.Metrics
}

func NewQueue(metrics *metrics.Metrics) *Queue {
	return &Queue{
		JobQueue: make(chan job.Job),
		Jobs:     make(map[string]job.Job),
		Metrics:  metrics,
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
