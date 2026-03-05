package main

import "github.com/google/uuid"

type WorkerState string

const (
	WorkerIdle    WorkerState = "Idle"
	WorkerBusy    WorkerState = "Busy"
	WorkerStopped WorkerState = "Stopped"
)

type Worker struct {
	ID       string
	State    WorkerState
	JobQueue chan Job
}

func NewWorker(jobQueue chan Job) Worker {
	return Worker{
		ID:       uuid.NewString(),
		State:    WorkerIdle,
		JobQueue: jobQueue,
	}
}
