package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type LoadBalancer struct {
	numUnicorns  int
	Unicorns     []*Unicorn
	algorithm    Algorithm
	responseChan chan *Response
	sendErrors   uint
	stopChan     chan bool
}

func NewLoadBalancer(numUnicorns int, algorithm Algorithm) *LoadBalancer {
	return &LoadBalancer{
		numUnicorns: numUnicorns,
		algorithm:   algorithm,
	}
}

func (lb *LoadBalancer) responseHandler() {
	for r := range lb.responseChan {
		lb.algorithm.ProcessResponse(r)
	}
}

func (lb *LoadBalancer) utilizationMonitor() {
	for {
		select {
		case <-lb.stopChan:
			break
		case <-time.After(1 * time.Second):
			utils := make([]float64, len(lb.Unicorns))
			var utils_sum float64
			var utils_squared_sum float64
			var overloaded int

			for i, u := range lb.Unicorns {
				utilization := u.Utilization()
				utils_sum += utilization
				utils[i] = utilization
			}

			avg := utils_sum / float64(len(utils))
			for _, u := range utils {
				utils_squared_sum += math.Pow(u-avg, 2)
			}
			variance := utils_squared_sum / float64(len(utils)-1)
			stddev := math.Sqrt(variance)

			overloaded_threshold := avg + (stddev * 2)
			for _, u := range utils {
				if u > overloaded_threshold {
					overloaded += 1
				}
			}
			overloaded_pct := float64(overloaded) / float64(len(utils))

			sort.Slice(utils, func(a, b int) bool { return utils[a] < utils[b] })
			p25 := utils[int(math.Floor(float64(len(utils))*0.25))]
			p75 := utils[int(math.Floor(float64(len(utils))*0.75))]
			p95 := utils[int(math.Floor(float64(len(utils))*0.95))]

			fmt.Printf("%s avg=%f std_dev=%f overloaded=%d overloaded_pct=%.3f p25=%f p75=%f p95=%f\n", time.Now().Format(time.RFC3339Nano), avg, stddev, overloaded, overloaded_pct, p25, p75, p95)
		}
	}
}

func (lb *LoadBalancer) Start() {
	lb.responseChan = make(chan *Response)
	lb.stopChan = make(chan bool)
	lb.Unicorns = make([]*Unicorn, lb.numUnicorns)

	for i := 0; i < lb.numUnicorns; i++ {
		unicorn := NewUnicorn(lb.responseChan)
		unicorn.Start()
		lb.Unicorns[i] = unicorn
	}

	go lb.responseHandler()
	go lb.utilizationMonitor()
}

func (lb *LoadBalancer) Stop() {
	close(lb.stopChan)
	for i := 0; i < lb.numUnicorns; i++ {
		lb.Unicorns[i].Stop()
	}
	close(lb.responseChan)
}

func (lb *LoadBalancer) Send(r *Request) {
	upstream := lb.algorithm.Select(lb.Unicorns)
	if !upstream.Send(r) {
		lb.sendErrors += 1
	}
}
