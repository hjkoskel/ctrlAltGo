/*
reboot
involves umounting safely and then reboot
*/
package initializing

import (
	"fmt"
	"syscall"
	"time"
)

func Reboot() error {
	mInfos, errInfos := GetMountInfo()
	if errInfos != nil {
		return fmt.Errorf("error getting mount info %s", errInfos)
	}
	syscall.Sync() //a little bit softer first
	time.Sleep(time.Millisecond * 250)
	mountpointList := mInfos.Mountpoints()
	fmt.Printf("\nMountinfos:%#v\n\nMOUNTPOINTS:%#v\n", mInfos, mountpointList)

	errUmount := UmountAll(mountpointList)
	if errUmount != nil {
		return errUmount
	}
	errReboot := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	if errReboot != nil {
		return errReboot
	}
	fmt.Printf("REBOOTING!!!\n")
	time.Sleep(time.Second * 10)
	fmt.Printf("FAIL, TIMEOUT ON REBOOT\n") //can this happen
	return fmt.Errorf("reboot timeout")
}
