package main

import (
	"bufio"
	"os"
	"strings"
	"time"
)

type RequestDispatcher struct {
	requests []*Request
}

func LoadRequestsFromFile(filename string) *RequestDispatcher {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	requests := make([]*Request, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		startTime, err := time.Parse(time.RFC3339Nano, strings.Trim(parts[0], "\""))
		if err != nil {
			continue
		}

		latency, err := time.ParseDuration(strings.Trim(parts[1], "\"") + "s")
		if err != nil {
			continue
		}

		request := &Request{
			OriginalStartTime: startTime,
			Latency:           latency,
		}
		requests = append(requests, request)
	}

	return &RequestDispatcher{
		requests: requests,
	}
}

func (rd *RequestDispatcher) NumRequests() int {
	return len(rd.requests)
}

func (rd *RequestDispatcher) Duration() time.Duration {
	return rd.requests[len(rd.requests)-1].OriginalStartTime.Sub(rd.requests[0].OriginalStartTime)
}

func (rd *RequestDispatcher) Execute(lb *LoadBalancer) {
	var lastRequest *Request
	for _, r := range rd.requests {
		if lastRequest != nil && r.OriginalStartTime != lastRequest.OriginalStartTime {
			time.Sleep(r.OriginalStartTime.Sub(lastRequest.OriginalStartTime))
		}
		lb.Send(r)
		lastRequest = r
	}
}
