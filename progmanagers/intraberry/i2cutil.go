/*
I2C utilities for debugging etc...
on real application tailor functionality or use thirparty library
*/
package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	I2C_SLAVE = 0x0703 // I2C slave address setting (from Linux I2C headers)
)

type I2CScanList []int

func ListI2CBusses() ([]string, error) {
	//Heuristics, does work on all hardwares
	contents, errReadDir := os.ReadDir("/dev")
	if errReadDir != nil {
		return nil, fmt.Errorf("error reading /dev err:%s", errReadDir)
	}

	result := []string{}
	for _, finfo := range contents {
		if finfo.IsDir() {
			continue
		}
		if strings.HasPrefix(finfo.Name(), "i2c-") {
			result = append(result, finfo.Name())
		}
	}
	return result, nil
}

// Scan addresses available and addresses in use/reserved
func ScanI2C(busPath string) (I2CScanList, I2CScanList, error) {
	// Open the I2C bus (e.g., /dev/i2c-1)
	//busPath := "/dev/i2c-1"
	file, err := os.OpenFile(busPath, os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open I2C bus %s: %v", busPath, err)

	}
	defer file.Close()

	lst := []int{}
	reservedList := []int{}

	// Scan all possible 7-bit I2C addresses (0x03 to 0x77)
	for addr := 0x03; addr <= 0x77; addr++ {
		_, _, errno := syscall.Syscall6(
			syscall.SYS_IOCTL,
			uintptr(file.Fd()),
			uintptr(I2C_SLAVE),
			uintptr(addr),
			0, 0, 0,
		)
		if errno != 0 {
			fmt.Printf(" [addr%d errno:%d=%s] ", addr, errno, errno.Error())
			if errno == syscall.EBUSY {
				reservedList = append(reservedList, addr)
			}
			continue
		}
		buf := make([]byte, 1)
		_, err := file.Read(buf) //Read back byte?
		if err == nil {
			//fmt.Printf("readback %d=>%d\n", addr, readback)
			lst = append(lst, addr)
		}
	}
	return lst, reservedList, nil
}

type I2CScanReport struct {
	Available map[int][]string `json:"available"`
	Reserved  map[int][]string `json:"reserved"`
}

func (p *I2CDeviceDatabase) CreateReport(lst I2CScanList, reservedLst I2CScanList) I2CScanReport {
	result := I2CScanReport{
		Available: make(map[int][]string),
		Reserved:  make(map[int][]string),
	}

	for _, addr := range lst {
		d := p.DevicesWithAddress(addr)
		result.Available[addr] = d.PartNumbers()
	}

	for _, addr := range reservedLst {
		d := p.DevicesWithAddress(addr)
		result.Reserved[addr] = d.PartNumbers()
	}

	return result
}
