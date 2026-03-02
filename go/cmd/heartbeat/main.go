package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seabit-ai/heartbeat/go/internal/collector"
	"github.com/seabit-ai/heartbeat/go/internal/config"
	"github.com/seabit-ai/heartbeat/go/internal/uploader"
)

// heartbeatEvent is the inner event payload sent to Splunk.
type heartbeatEvent struct {
	Event             string  `json:"event"`
	Host              string  `json:"host"`
	CPU               float64 `json:"cpu"`
	MemTotalMB        int64   `json:"memTotalMB"`
	MemUsedMB         int64   `json:"memUsedMB"`
	MemPercent        int     `json:"memPercent"`
	DiskUsedMB        int64   `json:"diskUsedMB"`
	DiskPercent       int     `json:"diskPercent"`
	RxKByte           int64   `json:"rxKByte"`
	TxKByte           int64   `json:"txKByte"`
	OS                string  `json:"os"`
	Arch              string  `json:"arch"`
	UptimeMinutes     int64   `json:"uptimeMinutes"`
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

	interval := time.Duration(cfg.HBIntervalSeconds) * time.Second

	log.Printf("heartbeat starting: host=%s interval=%s", hostname, interval)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first beat immediately
	if err := beat(cfg, hec, hostname); err != nil {
		log.Printf("beat error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := beat(cfg, hec, hostname); err != nil {
				log.Printf("beat error: %v", err)
			}
		case sig := <-stop:
			log.Printf("received signal %v, shutting down", sig)
			return
		}
	}
}

func beat(cfg *config.Config, hec *uploader.HECUploader, hostname string) error {
	// CPU: 1s sample
	cpu, err := collector.CPUPercent()
	if err != nil {
		log.Printf("cpu: %v", err)
	}
	cpu = roundTo1(cpu)

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

	now := float64(time.Now().UnixNano()) / 1e9

	inner := heartbeatEvent{
		Event:             "hostAgent",
		Host:              hostname,
		CPU:               cpu,
		MemTotalMB:        mem.TotalMB,
		MemUsedMB:         mem.UsedMB,
		MemPercent:        mem.PercentUsed,
		DiskUsedMB:        disk.UsedMB,
		DiskPercent:       disk.Percent,
		RxKByte:           net.RxKByte,
		TxKByte:           net.TxKByte,
		OS:                osInfo.OSName,
		Arch:              osInfo.Arch,
		UptimeMinutes:     uptimeMin,
		HBIntervalSeconds: cfg.HBIntervalSeconds,
	}

	evt := uploader.HECEvent{
		Time:   now,
		Host:   hostname,
		Source: "heartbeat",
		Index:  cfg.Index,
		Event:  inner,
	}

	log.Printf("sending heartbeat: cpu=%.1f%% mem=%dMB/%dMB disk=%dMB rx=%dKB tx=%dKB",
		cpu, mem.UsedMB, mem.TotalMB, disk.UsedMB, net.RxKByte, net.TxKByte)

	if cfg.HECURL == "" || cfg.HECToken == "" {
		fmt.Printf("(dry-run, no HEC configured)\n")
		return nil
	}

	return hec.Send(evt)
}

func roundTo1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
