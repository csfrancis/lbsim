package main

import (
	"sync"
	"sync/atomic"
	"time"
)

const unicornWorkers = 16
const unicornRequestBufferSize = 100

var unicornId = -1

type Unicorn struct {
	Id           int
	responseChan chan *Response
	requestChan  chan *Request
	workingCount int32
	waitGroup    *sync.WaitGroup
}

func NewUnicorn(responseChan chan *Response) *Unicorn {
	unicornId += 1
	return &Unicorn{
		Id:           unicornId,
		responseChan: responseChan,
	}
}

func (u *Unicorn) worker() {
	u.waitGroup.Add(1)
	defer u.waitGroup.Done()

	for r := range u.requestChan {
		atomic.AddInt32(&u.workingCount, 1)
		r.UnicornId = u.Id
		time.Sleep(r.Latency)
		r.EndTime = time.Now()
		atomic.AddInt32(&u.workingCount, -1)

		u.responseChan <- &Response{
			Request:     r,
			Utilization: u.Utilization(),
		}
	}
}

func (u *Unicorn) Start() {
	u.waitGroup = &sync.WaitGroup{}
	u.requestChan = make(chan *Request, unicornRequestBufferSize)

	for i := 0; i < unicornWorkers; i++ {
		go u.worker()
	}
}

func (u *Unicorn) Send(r *Request) bool {
	r.StartTime = time.Now()
	select {
	case u.requestChan <- r:
		return true
	default:
		return false
	}
}

func (u *Unicorn) Utilization() float64 {
	return float64(int(u.workingCount)+len(u.requestChan)) / float64(unicornWorkers)
}

func (u *Unicorn) Stop() {
	close(u.requestChan)
	u.waitGroup.Wait()
}
