//go:build !darwin

package collector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UptimeMinutes reads /proc/uptime and returns the uptime in minutes.
func UptimeMinutes() (int64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("read /proc/uptime: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected /proc/uptime format")
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	return int64(seconds / 60), nil
}
