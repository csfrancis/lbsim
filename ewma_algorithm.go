package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const decayTime = 10
const maxPeers = 500
const randomSeed = 93401285

const EWMALatency = 0
const EWMAUtilization = 1

type EWMAAlgorithm struct {
	lock         *sync.RWMutex
	scores       map[int]float64
	last_touched map[int]float64
	rand         *rand.Rand
	signal       int
}

func NewEWMAAlgorithm(signal int) *EWMAAlgorithm {
	switch signal {
	case EWMALatency:
	case EWMAUtilization:
		break
	default:
		panic("unknown EWMA signal")
	}

	scores := make(map[int]float64, maxPeers)
	last_touched := make(map[int]float64, maxPeers)
	for i := 0; i < maxPeers; i++ {
		scores[i] = 0
		last_touched[i] = 0
	}

	return &EWMAAlgorithm{
		lock:         &sync.RWMutex{},
		scores:       scores,
		last_touched: last_touched,
		rand:         rand.New(rand.NewSource(randomSeed)),
		signal:       signal,
	}
}

func timeToFloat64(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}

func decayWeight(last_touched float64, now time.Time) float64 {
	delta := timeToFloat64(time.Now()) - last_touched
	if delta < 0 {
		delta = 0
	}
	return math.Exp(-delta / decayTime)
}

func (e *EWMAAlgorithm) pickPeers(peers []*Unicorn) (peer1, peer2 *Unicorn) {
	size := int32(len(peers))
	idx1, idx2 := e.rand.Int31n(size), e.rand.Int31n(size)
	for idx1 == idx2 {
		idx2 = e.rand.Int31n(size)
	}
	return peers[idx1], peers[idx2]
}

func (e *EWMAAlgorithm) score(peerId int, now time.Time, update float64) float64 {
	ewma := e.scores[peerId] // empty values are 0
	last_touched := e.last_touched[peerId]

	weight := decayWeight(last_touched, now)
	ewma = ewma*weight + update*(1.0-weight)

	return ewma
}

func (e *EWMAAlgorithm) update(peerId int, val float64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	now := time.Now()
	ewma := e.score(peerId, now, val)

	e.last_touched[peerId] = timeToFloat64(now)
	e.scores[peerId] = ewma
}

func (e *EWMAAlgorithm) Select(peers []*Unicorn) *Unicorn {
	now := time.Now()
	p1, p2 := e.pickPeers(peers)

	e.lock.RLock()
	defer e.lock.RUnlock()

	if e.score(p1.Id, now, 0) < e.score(p2.Id, now, 0) {
		return p1
	} else {
		return p2
	}
}

func (e *EWMAAlgorithm) ProcessResponse(r *Response) {
	var signal float64
	if e.signal == EWMALatency {
		signal = float64(r.Request.Duration())
	} else if e.signal == EWMAUtilization {
		signal = r.Utilization
	}
	e.update(r.Request.UnicornId, signal)
}

func (e *EWMAAlgorithm) String() string {
	if e.signal == EWMALatency {
		return "EWMA(latency)"
	} else if e.signal == EWMAUtilization {
		return "EWMA(utilization)"
	}
	return "EWMA(unknown)"
}
