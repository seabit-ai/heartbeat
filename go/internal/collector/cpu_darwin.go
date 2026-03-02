package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CPUPercent on Darwin uses top -l 2 to get a CPU usage sample.
func CPUPercent() (float64, error) {
	out, err := exec.Command("sh", "-c", `top -l 2 -n 0 | grep "CPU usage" | tail -1`).Output()
	if err != nil {
		return 0, fmt.Errorf("top failed: %w", err)
	}
	// Line looks like: "CPU usage: 3.45% user, 5.12% sys, 91.42% idle"
	line := strings.TrimSpace(string(out))
	for _, field := range strings.Split(line, ",") {
		field = strings.TrimSpace(field)
		if strings.HasSuffix(field, "% idle") {
			idleStr := strings.TrimSuffix(field, "% idle")
			idleStr = strings.TrimSpace(idleStr)
			idle, err := strconv.ParseFloat(idleStr, 64)
			if err != nil {
				return 0, err
			}
			return 100.0 - idle, nil
		}
	}
	return 0, fmt.Errorf("could not parse top output: %q", line)
}
