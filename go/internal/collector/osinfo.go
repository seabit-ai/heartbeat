//go:build !darwin

package collector

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// OSInfo holds OS and architecture information.
type OSInfo struct {
	OSName string
	Arch   string
}

// GetOSInfo returns OS name and architecture.
func GetOSInfo() OSInfo {
	arch := runtime.GOARCH
	osName := readOSRelease()
	return OSInfo{OSName: osName, Arch: arch}
}

func readOSRelease() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}
	defer f.Close()

	vals := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], `"`)
		vals[key] = val
	}

	if pretty, ok := vals["PRETTY_NAME"]; ok && pretty != "" {
		return pretty
	}
	name := vals["NAME"]
	version := vals["VERSION"]
	if name != "" && version != "" {
		return name + " " + version
	}
	if name != "" {
		return name
	}
	return runtime.GOOS
}
