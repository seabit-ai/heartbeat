package collector

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetOSInfo on Darwin uses sw_vers to get the macOS version.
func GetOSInfo() OSInfo {
	return OSInfo{
		OSName:   darwinOSName(),
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
	}
}

func darwinOSName() string {
	prodName, err1 := exec.Command("sw_vers", "-productName").Output()
	prodVersion, err2 := exec.Command("sw_vers", "-productVersion").Output()
	if err1 != nil || err2 != nil {
		return fmt.Sprintf("macOS %s", runtime.GOARCH)
	}
	name := strings.TrimSpace(string(prodName))
	version := strings.TrimSpace(string(prodVersion))
	return fmt.Sprintf("%s %s", name, version)
}
