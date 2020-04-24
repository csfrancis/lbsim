package main

import (
	"time"
)

type Request struct {
	OriginalStartTime time.Time
	StartTime         time.Time
	EndTime           time.Time
	Latency           time.Duration
	UnicornId         int
}

func (r *Request) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}
