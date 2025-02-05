package initializing

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// MountInfo represents a single entry in /proc/mounts
type MountInfo struct {
	Device     string
	MountPoint string
	Filesystem string
	Options    string
	Free       SizeBytes
	Capacity   SizeBytes
}

type MountInfoArr []MountInfo

func (p *MountInfoArr) String() string {
	var sb strings.Builder
	for _, m := range *p {
		sb.WriteString(m.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

func (p *MountInfoArr) Devices() []string {
	result := make([]string, len(*p))
	for i, m := range *p {
		result[i] = m.Device
	}
	return result
}

func (p *MountInfoArr) Get(device string) *MountInfo { //Instead of using map
	for _, m := range *p {
		if device == m.Device {
			return &m
		}
	}
	return nil
}

func (a MountInfo) String() string {
	opts := ""
	if 0 < len(a.Options) {
		opts = "\nopts:" + a.Options
	}

	if a.Capacity == 0 {
		return fmt.Sprintf("%s: %s (%s) %s", a.Device, a.MountPoint, a.Filesystem,
			opts)
	}
	return fmt.Sprintf("%s: %s (%s) %s/%s = %.2f%% %s", a.Device, a.MountPoint, a.Filesystem,
		a.Free, a.Capacity, 100*float64(a.Free)/float64(a.Capacity),
		opts)
}

func GetMountInfo() (MountInfoArr, error) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("error reading /proc/mounts: %s", err)
	}
	lines := strings.Split(string(data), "\n")

	var mounts []MountInfo
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			mount := MountInfo{
				Device:     fields[0],
				MountPoint: fields[1],
				Filesystem: fields[2],
				Options:    fields[3],
			}

			var stat syscall.Statfs_t
			err := syscall.Statfs(mount.MountPoint, &stat)
			if err == nil {
				mount.Free = SizeBytes(int64(stat.Bavail) * stat.Bsize)
				mount.Capacity = SizeBytes(int64(stat.Blocks) * stat.Bsize)
			}

			mounts = append(mounts, mount)
		}
	}
	return mounts, nil
}
