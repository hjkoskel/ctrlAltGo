package initializing

import (
	"os"
	"syscall"
)

type MountCmd struct {
	Source string
	Target string
	FsType string
	Flags  uintptr
	Data   string
}

func (p *MountCmd) CreateAndMount() error {
	os.MkdirAll(p.Target, 0777)
	return syscall.Mount(p.Source, p.Target, p.FsType, p.Flags, p.Data)
}

func CreateAndMount(cmds []MountCmd) error {
	for _, cmd := range cmds {
		err := cmd.CreateAndMount()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
var MinimalMountCmds []MountCmd = []MountCmd{
	MountCmd{Source: "tmpfs", Target: "/tmp", FsType: "tmpfs", Flags: syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_RELATIME, Data: ""},
	MountCmd{Source: "devtmpfs", Target: "/dev", FsType: "devtmpfs", Flags: 0, Data: ""},
	MountCmd{Source: "devpts", Target: "/dev/pts", FsType: "devpts", Flags: 0, Data: ""},
	MountCmd{Source: "tmpfs", Target: "/dev/shm", FsType: "tmpfs", Flags: 0, Data: ""},
	MountCmd{Source: "tmpfs", Target: "/run", FsType: "tmpfs", Flags: 0, Data: ""},
	MountCmd{Source: "proc", Target: "/proc", FsType: "proc", Flags: 0, Data: ""},
	MountCmd{Source: "sysfs", Target: "/sys", FsType: "sysfs", Flags: 0, Data: ""},
}
*/

// MountNormal,
func MountNormal() error {
	os.Mkdir("/etc", 0777) //Does not contain mounts but have to be created as other important directories
	CreateAndMount([]MountCmd{
		{Source: "tmpfs", Target: "/tmp", FsType: "tmpfs", Flags: syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_RELATIME, Data: ""},
		{Source: "devtmpfs", Target: "/dev", FsType: "devtmpfs", Flags: 0, Data: ""},
		{Source: "devpts", Target: "/dev/pts", FsType: "devpts", Flags: 0, Data: ""},
		{Source: "tmpfs", Target: "/dev/shm", FsType: "tmpfs", Flags: 0, Data: ""},
		{Source: "tmpfs", Target: "/run", FsType: "tmpfs", Flags: 0, Data: ""}})

	busycmds := []MountCmd{
		{Source: "proc", Target: "/proc", FsType: "proc", Flags: 0, Data: ""},
		{Source: "sysfs", Target: "/sys", FsType: "sysfs", Flags: 0, Data: ""}}

	for _, m := range busycmds {
		err := m.CreateAndMount()
		if err == nil {
			continue
		}
		sce, ok := err.(syscall.Errno)
		if ok && sce == syscall.EBUSY {
			continue //ok to be busy in these mounts
		}
	}
	return nil
}
