#!/bin/bash
set -e

# Heartbeat installation script
# Usage: sudo ./install.sh

if [ "$EUID" -ne 0 ]; then
  echo "Error: This script must be run as root (use sudo)"
  exit 1
fi

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)
    ARCH_NAME="amd64"
    ;;
  aarch64|arm64)
    ARCH_NAME="arm64"
    ;;
  armv7l)
    ARCH_NAME="arm"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux)
    BINARY="dist/heartbeat-linux-$ARCH_NAME"
    ;;
  darwin)
    BINARY="dist/heartbeat-darwin-$ARCH_NAME"
    ;;
  *)
    echo "Error: Unsupported OS: $OS"
    exit 1
    ;;
esac

# Check if binary exists
if [ ! -f "$BINARY" ]; then
  echo "Error: Binary not found: $BINARY"
  echo "Run 'make build-all' first to build binaries"
  exit 1
fi

echo "Installing heartbeat for $OS/$ARCH_NAME..."

# Install binary
echo "  → Installing binary to /usr/local/bin/heartbeat"
cp "$BINARY" /usr/local/bin/heartbeat
chmod +x /usr/local/bin/heartbeat

# Create config directory
echo "  → Creating /etc/heartbeat/"
mkdir -p /etc/heartbeat

# Install config if not exists
if [ ! -f /etc/heartbeat/config.toml ]; then
  echo "  → Installing config.toml.example → /etc/heartbeat/config.toml"
  cp config.toml.example /etc/heartbeat/config.toml
  echo "     ⚠️  Edit /etc/heartbeat/config.toml to set hec_url and hec_token"
else
  echo "  → Config already exists: /etc/heartbeat/config.toml (not overwritten)"
fi

# Install service
if [ "$OS" = "linux" ]; then
  if [ -d /etc/systemd/system ]; then
    echo "  → Installing systemd service"
    cp systemd/heartbeat.service /etc/systemd/system/
    systemctl daemon-reload
    echo "     Service installed. Enable and start with:"
    echo "       sudo systemctl enable heartbeat"
    echo "       sudo systemctl start heartbeat"
  else
    echo "  ⚠️  systemd not found, skipping service installation"
  fi
elif [ "$OS" = "darwin" ]; then
  echo "  → Installing launchd service"
  cp launchd/com.seabit.heartbeat.plist /Library/LaunchDaemons/
  echo "     Service installed. Load with:"
  echo "       sudo launchctl load /Library/LaunchDaemons/com.seabit.heartbeat.plist"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Edit config: sudo vi /etc/heartbeat/config.toml"
echo "  2. Set hec_url and hec_token"
if [ "$OS" = "linux" ]; then
  echo "  3. Start service: sudo systemctl enable heartbeat && sudo systemctl start heartbeat"
  echo "  4. Check status: sudo systemctl status heartbeat"
elif [ "$OS" = "darwin" ]; then
  echo "  3. Load service: sudo launchctl load /Library/LaunchDaemons/com.seabit.heartbeat.plist"
  echo "  4. Check logs: sudo tail -f /var/log/heartbeat.log"
fi
