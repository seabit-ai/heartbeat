package collector

import (
	"syscall"
)

// DiskInfo holds disk usage statistics for the root filesystem.
type DiskInfo struct {
	UsedMB  int64
	Percent int
}

// DiskStats returns disk usage for "/".
func DiskStats() (DiskInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return DiskInfo{}, err
	}
	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bfree) * int64(stat.Bsize)
	used := total - free

	usedMB := used / 1024 / 1024
	percent := 0
	if total > 0 {
		percent = int(used * 100 / total)
	}
	return DiskInfo{UsedMB: usedMB, Percent: percent}, nil
}
