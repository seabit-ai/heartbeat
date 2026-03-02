package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// MemInfo holds memory statistics in MB.
type MemInfo struct {
	TotalMB     int64
	UsedMB      int64
	PercentUsed int
}

// MemStats on Darwin uses sysctl to read memory info.
func MemStats() (MemInfo, error) {
	// Total physical memory
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return MemInfo{}, fmt.Errorf("sysctl hw.memsize: %w", err)
	}
	totalBytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return MemInfo{}, err
	}
	totalMB := totalBytes / 1024 / 1024

	// Page size and page stats via vm_stat
	vmOut, err := exec.Command("vm_stat").Output()
	if err != nil {
		return MemInfo{}, fmt.Errorf("vm_stat: %w", err)
	}

	pageSize := int64(4096) // default
	pageSizeOut, err := exec.Command("sysctl", "-n", "hw.pagesize").Output()
	if err == nil {
		if ps, err2 := strconv.ParseInt(strings.TrimSpace(string(pageSizeOut)), 10, 64); err2 == nil {
			pageSize = ps
		}
	}

	var freePages, inactivePages int64
	for _, line := range strings.Split(string(vmOut), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Pages free:") {
			val := strings.Trim(strings.TrimPrefix(line, "Pages free:"), " .")
			freePages, _ = strconv.ParseInt(val, 10, 64)
		} else if strings.HasPrefix(line, "Pages inactive:") {
			val := strings.Trim(strings.TrimPrefix(line, "Pages inactive:"), " .")
			inactivePages, _ = strconv.ParseInt(val, 10, 64)
		}
	}

	availableBytes := (freePages + inactivePages) * pageSize
	usedMB := totalMB - availableBytes/1024/1024
	if usedMB < 0 {
		usedMB = 0
	}
	percent := 0
	if totalMB > 0 {
		percent = int(usedMB * 100 / totalMB)
	}
	return MemInfo{TotalMB: totalMB, UsedMB: usedMB, PercentUsed: percent}, nil
}
