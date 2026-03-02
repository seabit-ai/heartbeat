package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// NetStats holds network delta statistics in KB.
type NetStats struct {
	RxKByte int64
	TxKByte int64
}

type ifaceBytes struct {
	rx, tx uint64
}

func netstatSnapshot() (map[string]ifaceBytes, error) {
	out, err := exec.Command("netstat", "-ibn").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat -ibn: %w", err)
	}
	result := map[string]ifaceBytes{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		// Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
		if len(fields) < 10 {
			continue
		}
		iface := fields[0]
		if iface == "lo0" || strings.HasPrefix(iface, "Name") {
			continue
		}
		rx, err1 := strconv.ParseUint(fields[6], 10, 64)
		tx, err2 := strconv.ParseUint(fields[9], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		cur := result[iface]
		cur.rx += rx
		cur.tx += tx
		result[iface] = cur
	}
	return result, nil
}

// NetDelta samples netstat twice over intervalSeconds and returns KB delta.
func NetDelta(intervalSeconds int) (NetStats, error) {
	m1, err := netstatSnapshot()
	if err != nil {
		return NetStats{}, err
	}
	time.Sleep(time.Duration(intervalSeconds) * time.Second)
	m2, err := netstatSnapshot()
	if err != nil {
		return NetStats{}, err
	}

	var rxTotal, txTotal int64
	for iface, c2 := range m2 {
		c1 := m1[iface]
		if c2.rx > c1.rx {
			rxTotal += int64((c2.rx - c1.rx) / 1024)
		}
		if c2.tx > c1.tx {
			txTotal += int64((c2.tx - c1.tx) / 1024)
		}
	}
	return NetStats{RxKByte: rxTotal, TxKByte: txTotal}, nil
}
