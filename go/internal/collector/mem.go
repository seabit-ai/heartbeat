//go:build !darwin

package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MemInfo holds memory statistics in MB.
type MemInfo struct {
	TotalMB   int64
	UsedMB    int64
	PercentUsed int
}

// MemStats reads /proc/meminfo and returns memory usage.
func MemStats() (MemInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemInfo{}, err
	}
	defer f.Close()

	vals := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		vals[key] = v
	}

	total, ok1 := vals["MemTotal"]
	free, ok2 := vals["MemFree"]
	buffers, ok3 := vals["Buffers"]
	cached, ok4 := vals["Cached"]
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return MemInfo{}, fmt.Errorf("missing keys in /proc/meminfo")
	}

	// All values are in kB
	totalMB := int64(total / 1024)
	available := free + buffers + cached
	usedMB := int64((total - available) / 1024)
	if usedMB < 0 {
		usedMB = 0
	}
	percent := 0
	if totalMB > 0 {
		percent = int(usedMB * 100 / totalMB)
	}
	return MemInfo{TotalMB: totalMB, UsedMB: usedMB, PercentUsed: percent}, nil
}
