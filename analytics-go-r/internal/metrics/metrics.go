// Package metrics tracks lightweight in-process counters for the
// ingestion pipeline so the risk desk dashboard can show throughput
// without standing up a full metrics backend for this demo.
package metrics

import "sync"

// Counters tracks a fixed set of named counters.
type Counters struct {
	mu     sync.Mutex
	values map[string]int64
}

// NewCounters builds an empty Counters set.
func NewCounters() *Counters {
	return &Counters{values: make(map[string]int64)}
}

// Increment adds delta to the named counter.
func (c *Counters) Increment(name string, delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[name] += delta
}

// Snapshot returns a copy of all counters for reporting.
func (c *Counters) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	snapshot := make(map[string]int64, len(c.values))
	for name, value := range c.values {
		snapshot[name] = value
	}
	return snapshot
}
