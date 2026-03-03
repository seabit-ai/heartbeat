package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/seabit-ai/heartbeat/go/internal/collector"
	"github.com/seabit-ai/heartbeat/go/internal/config"
	"github.com/seabit-ai/heartbeat/go/internal/uploader"
)

// heartbeatEvent is the inner event payload sent to Splunk.
// Field names match the original Node.js heartbeat for Splunk compatibility.
type heartbeatEvent struct {
	Event             string  `json:"event"`
	Host              string  `json:"host"`
	Arch              string  `json:"arch"`
	UptimeMinutes     int64   `json:"uptimeMinutes"`
	MemTotalMB        int64   `json:"memTotalMB"`
	MemPercent        int     `json:"memPercent"`
	MemMB             int64   `json:"memMB"`
	CPU               float64 `json:"cpu"`
	CPUDetail         string  `json:"cpuDetail,omitempty"`
	CPUCount          int     `json:"cpuCount"`
	OS                string  `json:"os"`
	RxKByte           int64   `json:"rxKByte"`
	TxKByte           int64   `json:"txKByte"`
	DiskUsedMB        int64   `json:"diskUsedMB"`
	DiskPercent       int     `json:"diskPercent"`
	HBIntervalSeconds int     `json:"hbIntervalSeconds"`
}

func main() {
	configPath := flag.String("config", "config.toml", "path to config.toml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	hostname := cfg.Host
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			// Use short hostname only (strip domain suffix)
			if idx := strings.Index(h, "."); idx > 0 {
				h = h[:idx]
			}
			hostname = h
		} else {
			hostname = "unknown"
		}
	}

	if cfg.HECURL == "" {
		log.Println("WARNING: hec_url not configured, events will not be uploaded")
	}
	if cfg.HECToken == "" {
		log.Println("WARNING: hec_token not configured, events will not be uploaded")
	}

	hec := uploader.New(cfg.HECURL, cfg.HECToken)

	// Start background CPU sampler (every cpu_detail_interval_seconds, default 10s)
	sampler := &collector.CpuSampler{}
	sampler.Start(cfg.CPUDetailIntervalSeconds)

	interval := time.Duration(cfg.HBIntervalSeconds) * time.Second
	log.Printf("heartbeat starting: host=%s interval=%s cpuSampleInterval=%ds",
		hostname, interval, cfg.CPUDetailIntervalSeconds)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first beat immediately
	if err := beat(cfg, hec, hostname, sampler); err != nil {
		log.Printf("beat error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := beat(cfg, hec, hostname, sampler); err != nil {
				log.Printf("beat error: %v", err)
			}
		case sig := <-stop:
			log.Printf("received signal %v, shutting down", sig)
			return
		}
	}
}

func beat(cfg *config.Config, hec *uploader.HECUploader, hostname string, sampler *collector.CpuSampler) error {
	// CPU: from background sampler (avg + detail string)
	cpu, cpuDetail := sampler.GetAndReset()

	// Memory
	mem, err := collector.MemStats()
	if err != nil {
		log.Printf("mem: %v", err)
	}

	// Disk
	disk, err := collector.DiskStats()
	if err != nil {
		log.Printf("disk: %v", err)
	}

	// Network delta (1s)
	net, err := collector.NetDelta(1)
	if err != nil {
		log.Printf("net: %v", err)
	}

	// OS info
	osInfo := collector.GetOSInfo()

	// Uptime
	uptimeMin, err := collector.UptimeMinutes()
	if err != nil {
		log.Printf("uptime: %v", err)
	}

	now := time.Now().Unix()

	inner := heartbeatEvent{
		Event:             "hostAgent",
		Host:              hostname,
		Arch:          osInfo.Arch,
		UptimeMinutes: uptimeMin,
		MemTotalMB:        mem.TotalMB,
		MemPercent:        mem.PercentUsed,
		MemMB:             mem.UsedMB,
		CPU:               cpu,
		CPUDetail:         cpuDetail,
		CPUCount:          osInfo.CPUCount,
		OS:                osInfo.OSName,
		RxKByte:           net.RxKByte,
		TxKByte:           net.TxKByte,
		DiskUsedMB:        disk.UsedMB,
		DiskPercent:       disk.Percent,
		HBIntervalSeconds: cfg.HBIntervalSeconds,
	}

	evt := uploader.HECEvent{
		Time:   now,
		Host:   hostname,
		Source: "heartbeat",
		Index:  cfg.Index,
		Event:  inner,
	}

	log.Printf("event=sendHeartbeat cpu=%.1f cpuSamples=%d memMB=%d memTotalMB=%d diskUsedMB=%d diskPercent=%d rxKByte=%d txKByte=%d uptimeMinutes=%d",
		cpu, len(strings.Split(cpuDetail, ",")), mem.UsedMB, mem.TotalMB, disk.UsedMB, disk.Percent, net.RxKByte, net.TxKByte, uptimeMin)

	if cfg.HECURL == "" || cfg.HECToken == "" {
		if b, err := json.MarshalIndent(evt, "", "  "); err == nil {
			fmt.Printf("(dry-run) payload:\n%s\n", string(b))
		}
		return nil
	}

	return hec.Send(evt)
}
