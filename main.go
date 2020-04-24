package main

import (
	"flag"
	"fmt"
	"math"
)

const defaultUnicornUpstreams = 35
const defaultRequestsFilename = "lb_requests.csv"

func main() {
	algorithm := flag.String("algo", "round_robin", "algorithm to use: round_robin, perfect, ewma_latency, ewma_util")
	unicornUpstreams := flag.Int("unicorns", defaultUnicornUpstreams, "number of unicorn upstreams")
	requestsFilename := flag.String("file", defaultRequestsFilename, "request CSV file")

	flag.Parse()

	var algo Algorithm
	switch *algorithm {
	case "round_robin":
		algo = &RoundRobinAlgorithm{}
	case "perfect":
		algo = &PerfectAlgorithm{}
	case "ewma_latency":
		algo = NewEWMAAlgorithm(EWMALatency)
	case "ewma_util":
		algo = NewEWMAAlgorithm(EWMAUtilization)
	}

	lb := NewLoadBalancer(*unicornUpstreams, algo)
	dispatcher := LoadRequestsFromFile(*requestsFilename)

	fmt.Printf("Starting %d Unicorns and executing %d requests (~%.f seconds) using %v...\n", *unicornUpstreams, dispatcher.NumRequests(), math.Round(dispatcher.Duration().Seconds()), algo)
	lb.Start()
	dispatcher.Execute(lb)
	fmt.Printf("Request execution completed. Shutting down ... ")

	lb.Stop()
	fmt.Printf("done.\n")
}
