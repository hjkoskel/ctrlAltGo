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

	ctrlaltgo.JamIfErr(fmt.Errorf("end of program"))
}
