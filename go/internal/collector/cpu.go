//go:build !darwin

package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// cpuTimes holds the raw CPU time counters from /proc/stat.
type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func readCPUTimes() (cpuTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			return cpuTimes{}, fmt.Errorf("unexpected /proc/stat format")
		}
		var t cpuTimes
		vals := []*uint64{&t.user, &t.nice, &t.system, &t.idle, &t.iowait, &t.irq, &t.softirq, &t.steal}
		for i, v := range vals {
			n, err := strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return cpuTimes{}, err
			}
			*v = n
		}
		return t, nil
	}
	return cpuTimes{}, fmt.Errorf("cpu line not found in /proc/stat")
}

// CPUPercent samples /proc/stat twice over 1 second and returns the usage %.
func CPUPercent() (float64, error) {
	t1, err := readCPUTimes()
	if err != nil {
		return 0, err
	}
	time.Sleep(time.Second)
	t2, err := readCPUTimes()
	if err != nil {
		return 0, err
	}

	idle1 := t1.idle + t1.iowait
	idle2 := t2.idle + t2.iowait

	total1 := t1.user + t1.nice + t1.system + t1.idle + t1.iowait + t1.irq + t1.softirq + t1.steal
	total2 := t2.user + t2.nice + t2.system + t2.idle + t2.iowait + t2.irq + t2.softirq + t2.steal

	totalDiff := total2 - total1
	idleDiff := idle2 - idle1

	if totalDiff == 0 {
		return 0, nil
	}
	return 100.0 * float64(totalDiff-idleDiff) / float64(totalDiff), nil
}
