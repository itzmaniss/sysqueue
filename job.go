package main

import (
	"time"

	"github.com/google/uuid"
)

type JobState string

const (
	StateQueued  JobState = "queued"
	StateRunning JobState = "running"
	StateDone    JobState = "done"
	StateFailed  JobState = "failed"
)

type Job struct {
	ID        string
	Command   string
	State     JobState
	PID       *int
	StartedAt time.Time
	Duration  time.Duration
}

func NewJob(command string) Job {
	return Job{
		ID:      uuid.NewString(),
		Command: command,
		State:   StateQueued,
	}
}
