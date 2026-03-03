# heartbeat

Host metrics agent — collects CPU/mem/disk/network every 60s and posts to Splunk HEC.

**Go rewrite** (2026) — single binary, direct Splunk HEC upload, cross-platform.  
**Original Node.js version** preserved in `lib/`, `index.js` (legacy).

## Quick Start

```bash
# 1. Clone and build
git clone https://github.com/seabit-ai/heartbeat.git
cd heartbeat
make build-all   # or make linux-amd64, make darwin-arm64, etc.

# 2. Install
sudo ./install.sh

# 3. Configure
sudo vi /etc/heartbeat/config.toml   # set hec_url and hec_token

# 4. Start service
# Linux:
sudo systemctl enable heartbeat && sudo systemctl start heartbeat

# macOS:
sudo launchctl load /Library/LaunchDaemons/com.seabit.heartbeat.plist
```

The `install.sh` script auto-detects your platform and installs the correct binary.

## Heartbeat Payload

Posted to Splunk HEC `/services/collector/event`:

```json
{
  "time": 1772508432,
  "host": "m3u",
  "source": "heartbeat",
  "index": "heartbeat",
  "event": {
    "event": "hostAgent",
    "host": "m3u",
    "arch": "arm64",
    "uptimeMinutes": 1812,
    "memTotalMB": 98304,
    "memPercent": 32,
    "memMB": 31758,
    "cpu": 4.7,
    "cpuDetail": "3.8,4.2,5.1,4.9,5.2,4.3",
    "cpuCount": 28,
    "os": "macOS 26.3",
    "rxKByte": 26,
    "txKByte": 28,
    "diskUsedMB": 127568,
    "diskPercent": 6,
    "hbIntervalSeconds": 60
  }
}
```

## Configuration

`config.toml`:
```toml
hec_url = "https://splunk.example.com/services/collector/event"
hec_token = "YOUR-HEC-TOKEN-HERE"
hb_interval_seconds = 60
cpu_detail_interval_seconds = 10   # CPU sampling interval for cpuDetail field
host = ""                          # auto-detect if empty
index = "heartbeat"
```

## Deployment (Linux)

```bash
# 1. Copy binary
sudo cp dist/heartbeat-linux-amd64 /usr/local/bin/heartbeat
sudo chmod +x /usr/local/bin/heartbeat

# 2. Create config
sudo mkdir -p /etc/heartbeat
sudo cp config.toml /etc/heartbeat/

# 3. Install systemd service
sudo cp systemd/heartbeat.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable heartbeat
sudo systemctl start heartbeat

# 4. Check status
sudo systemctl status heartbeat
sudo journalctl -u heartbeat -f
```

## Build Targets

| Binary | Platform | Use Case |
|--------|----------|----------|
| `heartbeat-linux-amd64` | x86_64 Linux | DigitalOcean droplets, cloud VMs |
| `heartbeat-linux-arm64` | ARM64 Linux | NanoPi M6, Raspberry Pi 4/5 |
| `heartbeat-linux-arm` | ARMv7 Linux | Raspberry Pi 2/3, older boards |
| `heartbeat-darwin-arm64` | macOS M-series | Mac Studio, MacBook Pro M1+ |

Cross-compile with `make build-all` — produces all 4 binaries in `dist/`.

## Development

```bash
# Run locally (dry-run mode if hec_url empty)
go run ./go/cmd/heartbeat

# Build
go build -o heartbeat ./go/cmd/heartbeat

# Test
go test ./...
```

## Legacy Node.js Version

Original Node.js implementation preserved in:
- `lib/` — collectors (cpu, mem, disk, net, os)
- `index.js` — entry point
- See git history before 2026-02-24 for usage docs

The Node.js version required a separate `rain` service as middleware.  
The Go rewrite posts directly to Splunk HEC, eliminating the extra hop.

## License

MIT
