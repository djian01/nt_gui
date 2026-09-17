package testengine

import "container/heap"

// Each exact nearest-rank percentile partitions all successful RTTs into two
// heaps. Updates cost O(log n), with O(n) floats and no per-probe history scan.
type rttHeap struct {
	values  []float64
	maximum bool
}

func (h rttHeap) Len() int { return len(h.values) }
func (h rttHeap) Less(i, j int) bool {
	if h.maximum {
		return h.values[i] > h.values[j]
	}
	return h.values[i] < h.values[j]
}
func (h rttHeap) Swap(i, j int) { h.values[i], h.values[j] = h.values[j], h.values[i] }
func (h *rttHeap) Push(v any)   { h.values = append(h.values, v.(float64)) }
func (h *rttHeap) Pop() any {
	i := len(h.values) - 1
	v := h.values[i]
	h.values = h.values[:i]
	return v
}

type rankPercentile struct{ lower, upper rttHeap }

func (p *rankPercentile) add(value float64, percent int) float64 {
	p.lower.maximum = true
	if p.lower.Len() == 0 || value <= p.lower.values[0] {
		heap.Push(&p.lower, value)
	} else {
		heap.Push(&p.upper, value)
	}
	n := p.lower.Len() + p.upper.Len()
	rank := (n/100)*percent + ((n%100)*percent+99)/100
	for p.lower.Len() > rank {
		heap.Push(&p.upper, heap.Pop(&p.lower))
	}
	for p.lower.Len() < rank {
		heap.Push(&p.lower, heap.Pop(&p.upper))
	}
	return p.lower.values[0]
}

type latencyPercentiles struct{ p95, p99 rankPercentile }

func (p *latencyPercentiles) add(sample Sample, session *Session) {
	if !sample.Success {
		return
	}
	p95, p99 := p.p95.add(sample.RTT, 95), p.p99.add(sample.RTT, 99)
	// New values keep previously emitted Session snapshots immutable.
	session.P95RTT, session.P99RTT = &p95, &p99
}
