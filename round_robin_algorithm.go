package main

type RoundRobinAlgorithm struct {
	index int
}

func (rr *RoundRobinAlgorithm) Select(peers []*Unicorn) *Unicorn {
	rr.index += 1
	if rr.index >= len(peers) {
		rr.index = 0
	}
	return peers[rr.index]
}

func (rr *RoundRobinAlgorithm) ProcessResponse(_ *Response) {}

func (rr *RoundRobinAlgorithm) String() string {
	return "RoundRobinAlgorithm"
}
