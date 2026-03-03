package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// CpuSampler samples CPU usage in the background and accumulates values
// between heartbeats, matching the original Node.js cpuDetail behavior.
type CpuSampler struct {
	mu      sync.Mutex
	samples []float64
}

// Start begins background sampling every intervalSeconds.
// Call this once at startup; runs until the process exits.
func (s *CpuSampler) Start(intervalSeconds int) {
	go func() {
		for {
			v, err := CPUPercent()
			if err == nil {
				v = roundTo1(v)
				s.mu.Lock()
				s.samples = append(s.samples, v)
				s.mu.Unlock()
			}
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
	}()
}

// GetAndReset returns the average CPU% and cpuDetail string, then clears the buffer.
// Returns (0, "") if no samples yet.
func (s *CpuSampler) GetAndReset() (avg float64, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.samples) == 0 {
		return 0, ""
	}

	parts := make([]string, len(s.samples))
	sum := 0.0
	for i, v := range s.samples {
		sum += v
		parts[i] = fmt.Sprintf("%.1f", v)
	}
	avg = roundTo1(sum / float64(len(s.samples)))
	detail = strings.Join(parts, ",")
	s.samples = s.samples[:0]
	return avg, detail
}
