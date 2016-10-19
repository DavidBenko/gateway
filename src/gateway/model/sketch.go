package model

import (
	"sync/atomic"
	"time"
)

var (
	jobsExecuted int64
	start        = time.Now()
)

type Sketch struct {
	JobsExecuted int64 `json:"jobs_executed"`
	Uptime       int64 `json:"uptime"`
}

func NewSketch() *Sketch {
	return &Sketch{
		JobsExecuted: atomic.LoadInt64(&jobsExecuted),
		Uptime:       int64(time.Now().Sub(start) / time.Second),
	}
}

func IncrementJobsExecuted() {
	atomic.AddInt64(&jobsExecuted, 1)
}
