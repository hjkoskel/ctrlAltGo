/*
Simple example how to create single executable as init program and deploy that to various platform

just print some information and stops

*/

package main

import (
	"ctrlaltgo"
	"fmt"
	"initializing"
	"os"
	"time"
)

func listDir(dirname string) error {
	entries, errRead := os.ReadDir(dirname)
	if errRead != nil {
		return errRead
	}

	if errRead != nil {
		return errRead
	}

	fmt.Printf("DIR: %s  ", dirname)
	for _, entry := range entries {
		fmt.Printf("%s", entry.Name())
		if entry.IsDir() {
			fmt.Printf("(D)  ")
		} else {
			fmt.Printf("(F)  ")
		}
	}
	fmt.Printf("\n\n")
	return nil
}

func checkExampleBinDirFile() {
	content, err := os.ReadFile("/bin/example.txt")
	if err == nil {
		fmt.Printf("HEY! have example file: %s\n", content)
		return
	}

}

func main() {
	if os.Getpid() != 1 {
		fmt.Printf("This is initramfs program. Please run this as PID1\n")
		return
	}
	fmt.Printf("-- Starting example init program --\n")

	fmt.Printf("mounting normal set of dirs\n")
	errMount := initializing.MountNormal()
	ctrlaltgo.JamIfErr(errMount)
	listDir("/dev")

	for {
		fmt.Printf("--Block devices poll---\n")

		blockDevices, errBlockDevices := initializing.GetBlockDevices()
		ctrlaltgo.JamIfErr(errBlockDevices)

		// Print block device information
		/*
			for _, device := range blockDevices {
				if device.IsUSB {
					fmt.Printf("Device: %s (USB)\n", device.Name)
				} else {
					fmt.Printf("Device: %s\n", device.Name)
				}
				fmt.Printf("  Size: %d Vendor: %s Model: %s  Partitions: %v\n", device.SizeBytes, device.Vendor, device.Model, len(device.Partitions))
				for i, part := range device.Partitions {
					fmt.Printf("		%v:%s size:%v fstype:%s UUID:%s\n", i, part.Name, part.SizeGB, part.FSType, part.UUID)
				}

				fmt.Printf("  Is USB: %v\n", device.IsUSB)
				fmt.Println()
			}*/
		fmt.Printf("%s\n", blockDevices)

		mntInfo, errMntInfo := initializing.GetMountInfo()
		ctrlaltgo.JamIfErr(errMntInfo)
		fmt.Printf("\n")
		for _, m := range mntInfo {
			fmt.Printf("%s\n", m)
		}

		fmt.Printf("------ROOT--------\n")
		listDir("/")
		fmt.Printf("------BIN--------\n")
		listDir("/bin")
		checkExampleBinDirFile()
		time.Sleep(time.Second * 3)
	}
	ctrlaltgo.JamIfErr(fmt.Errorf("end of program"))
}
