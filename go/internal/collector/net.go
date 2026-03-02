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

// NetStats holds network delta statistics in KB.
type NetStats struct {
	RxKByte int64
	TxKByte int64
}

type ifaceCounters struct {
	rx, tx uint64
}

func readNetDev() (map[string]ifaceCounters, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := map[string]ifaceCounters{}
	scanner := bufio.NewScanner(f)
	// Skip 2 header lines
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		iface := strings.TrimSpace(line[:colon])
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 9 {
			continue
		}
		rx, err1 := strconv.ParseUint(fields[0], 10, 64)
		tx, err2 := strconv.ParseUint(fields[8], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		result[iface] = ifaceCounters{rx: rx, tx: tx}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no interfaces found in /proc/net/dev")
	}
	return result, nil
}

func sumCounters(m map[string]ifaceCounters) (rx, tx uint64) {
	for iface, c := range m {
		// Skip loopback
		if iface == "lo" {
			continue
		}
		rx += c.rx
		tx += c.tx
	}
	return
}

// NetDelta samples /proc/net/dev twice over intervalSeconds and returns KB delta.
func NetDelta(intervalSeconds int) (NetStats, error) {
	m1, err := readNetDev()
	if err != nil {
		return NetStats{}, err
	}
	time.Sleep(time.Duration(intervalSeconds) * time.Second)
	m2, err := readNetDev()
	if err != nil {
		return NetStats{}, err
	}

	rx1, tx1 := sumCounters(m1)
	rx2, tx2 := sumCounters(m2)

	rxDelta := int64(0)
	txDelta := int64(0)
	if rx2 > rx1 {
		rxDelta = int64((rx2 - rx1) / 1024)
	}
	if tx2 > tx1 {
		txDelta = int64((tx2 - tx1) / 1024)
	}
	return NetStats{RxKByte: rxDelta, TxKByte: txDelta}, nil
}
