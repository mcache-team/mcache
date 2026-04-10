package cluster

import (
	"sync"
	"time"
)

var DefaultTelemetry = newTelemetry()

var writeLatencyBuckets = []time.Duration{
	time.Millisecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	25 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
	time.Second,
	2500 * time.Millisecond,
	5 * time.Second,
	10 * time.Second,
}

type telemetry struct {
	mu         sync.RWMutex
	writeStats map[string]*writeStat
}

type writeStat struct {
	Requests            uint64
	Successes           uint64
	Errors              uint64
	Redirects           uint64
	LatencyTotalNs      uint64
	LastLatencyNs       uint64
	LatencyBucketCounts []uint64
	LatencyInfBucket    uint64
}

type TelemetrySnapshot struct {
	Writes map[string]writeStat `json:"writes"`
}

func newTelemetry() *telemetry {
	return &telemetry{
		writeStats: make(map[string]*writeStat),
	}
}

func (t *telemetry) ObserveWrite(operation string, duration time.Duration, outcome string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	stat, ok := t.writeStats[operation]
	if !ok {
		stat = &writeStat{
			LatencyBucketCounts: make([]uint64, len(writeLatencyBuckets)),
		}
		t.writeStats[operation] = stat
	}

	stat.Requests++
	stat.LatencyTotalNs += uint64(duration)
	stat.LastLatencyNs = uint64(duration)
	stat.LatencyInfBucket++

	for index, bucket := range writeLatencyBuckets {
		if duration <= bucket {
			stat.LatencyBucketCounts[index]++
		}
	}

	switch outcome {
	case "success":
		stat.Successes++
	case "redirect":
		stat.Redirects++
	case "error":
		stat.Errors++
	}
}

func (t *telemetry) Snapshot() TelemetrySnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	snapshot := TelemetrySnapshot{
		Writes: make(map[string]writeStat, len(t.writeStats)),
	}
	for operation, stat := range t.writeStats {
		copyStat := *stat
		copyStat.LatencyBucketCounts = append([]uint64(nil), stat.LatencyBucketCounts...)
		snapshot.Writes[operation] = copyStat
	}
	return snapshot
}
