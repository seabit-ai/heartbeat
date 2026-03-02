package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// UptimeMinutes on Darwin uses sysctl kern.boottime to compute uptime.
func UptimeMinutes() (int64, error) {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0, fmt.Errorf("sysctl kern.boottime: %w", err)
	}
	// Output: "{ sec = 1700000000, usec = 123456 } Mon Jan 01 00:00:00 2024"
	line := strings.TrimSpace(string(out))
	// Find "sec = <number>"
	const prefix = "sec = "
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return 0, fmt.Errorf("cannot parse kern.boottime: %q", line)
	}
	rest := line[idx+len(prefix):]
	// rest is "<number>, ..."
	end := strings.IndexAny(rest, ", }")
	if end < 0 {
		end = len(rest)
	}
	bootSec, err := strconv.ParseInt(strings.TrimSpace(rest[:end]), 10, 64)
	if err != nil {
		return 0, err
	}
	uptimeSec := time.Now().Unix() - bootSec
	return uptimeSec / 60, nil
}
