package main

import "sync"

type Queue struct {
	JobQueue chan Job
	Workers  map[string]Worker
	Jobs     map[string]Job
	Lock     sync.Mutex
}

func NewQueue(workerCount int) *Queue {
	workers := make(map[string]Worker)
	jobQueue := make(chan Job)
	jobsMap := make(map[string]Job)
	q := Queue{
		JobQueue: jobQueue,
		Jobs:     jobsMap,
	}

	for range workerCount {
		w := NewWorker(&q)
		workers[w.ID] = w
	}

	q.Workers = workers

	return &q
}

func (q *Queue) Enqueue(command string) {
	job := NewJob(command)
	q.Lock.Lock()
	q.Jobs[job.ID] = job
	q.Lock.Unlock()
	q.JobQueue <- job
}
