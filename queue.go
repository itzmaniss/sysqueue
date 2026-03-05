package main

type Queue struct {
	JobQueue chan Job
	Workers  map[string]Worker
	Jobs     map[string]Job
}

func NewQueue(workerCount int) Queue {
	workers := make(map[string]Worker)
	jobQueue := make(chan Job)
	for range workerCount {
		w := NewWorker(jobQueue)
		workers[w.ID] = w
	}
	return Queue{
		JobQueue: jobQueue,
		Workers:  workers,
		Jobs:     make(map[string]Job),
	}
}
