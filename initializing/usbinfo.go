/*
This package is for detecting usb devices on initialization
*/
package initializing

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	USBDEVICESDIR = "/sys/bus/usb/devices"
)

func readOneHexNumberFile(fname string) (int64, error) {
	byt, err := os.ReadFile(fname)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(byt)), 16, 64)
}

// UsbVidPidFound  check is device connected
func UsbVidPidFound(vid uint16, pid uint16) (bool, error) {
	dirlist, errDirlist := os.ReadDir(USBDEVICESDIR)
	if errDirlist != nil {
		return false, errDirlist
	}
	for _, entry := range dirlist {

		idvendorfilename, errEvalVendorIdPath := filepath.EvalSymlinks(path.Join(USBDEVICESDIR, entry.Name(), "idVendor"))
		if errEvalVendorIdPath != nil {
			continue
		}
		idVendor, errIdVendor := readOneHexNumberFile(idvendorfilename)
		if errIdVendor != nil {
			return false, errIdVendor
		}

		idproductfilename, errEvalProductIdPath := filepath.EvalSymlinks(path.Join(USBDEVICESDIR, entry.Name(), "idProduct"))
		if errEvalProductIdPath != nil {
			//fmt.Printf("no idProduct %s\n", errEvalProductIdPath)
			continue
			//return false, errEvalProductIdPath
		}
		idProduct, errIdProduct := readOneHexNumberFile(idproductfilename)
		if errIdProduct != nil {
			continue
		}

		//fmt.Printf("%s VID:%X, PID:%X\n", entry.Name(), idVendor, idProduct)
		if vid == uint16(idVendor) && pid == uint16(idProduct) {
			return true, nil
		}
	}
	return false, nil
}
