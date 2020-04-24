package main

type PerfectAlgorithm struct {
}

func (p *PerfectAlgorithm) Select(peers []*Unicorn) *Unicorn {
	var peer *Unicorn
	for _, u := range peers {
		if peer == nil {
			peer = u
		} else if u.Utilization() < peer.Utilization() {
			peer = u
		}
	}
	return peer
}

func (p *PerfectAlgorithm) ProcessResponse(_ *Response) {}

func (p *PerfectAlgorithm) String() string {
	return "PerfectAlgorithm"
}
