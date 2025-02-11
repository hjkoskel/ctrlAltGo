/*
common utilities... TODO separate sub repo? TODO or move initializing here
*/
package ctrlaltgo

import (
	"debug/elf"
	"fmt"
	"os"
	"time"
)

// JamIfErr Prints error and jams so kernel panic is not caused. USE on development only. Find better way to report error and recover from it
func JamIfErr(err error) {
	if err == nil {
		return
	}
	fmt.Printf("\n%s\n\n", err)
	for err != nil {
		if os.Getpid() != 1 {
			os.Exit(-1)
		}
		time.Sleep(time.Second * 5)
	}
}

// GetCurrentMachine returns the machine type of the currently running binary.
func GetCurrentMachine() (elf.Machine, error) {
	// Get the path to the currently running binary
	exePath, err := os.Executable()
	if err != nil {
		return elf.EM_NONE, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Open the ELF file
	file, err := elf.Open(exePath)
	if err != nil {
		return elf.EM_NONE, fmt.Errorf("failed to open ELF file: %w", err)
	}
	defer file.Close()

	// Return the machine type
	return file.Machine, nil
}
