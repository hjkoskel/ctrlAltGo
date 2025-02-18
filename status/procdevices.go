/*
lookup of /proc/devices
*/
package status

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ProcDevices struct {
	BlockDevices     map[int64]string //Major device numbers
	CharacterDevices map[int64]string
}

func LoadProcDevices() (ProcDevices, error) {
	byt, err := os.ReadFile("/proc/devices")
	if err != nil {
		return ProcDevices{}, err
	}
	rows := strings.Split(string(byt), "\n")
	result := ProcDevices{
		BlockDevices:     make(map[int64]string),
		CharacterDevices: make(map[int64]string),
	}

	readBlock := false
	readChar := false
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) == 0 {
			continue
		}
		if row == "Character devices:" {
			readChar = true
			readBlock = false
			continue
		}
		if row == "Block devices:" {
			readChar = true
			readBlock = false
			continue
		}
		if readChar == readBlock && !readBlock {
			return ProcDevices{}, fmt.Errorf("failed loading %#v", rows)
		}
		fie := strings.Fields(row)
		if len(fie) != 2 {
			return ProcDevices{}, fmt.Errorf("unexpected number of columns %#v", fie)
		}
		major, errParse := strconv.ParseInt(fie[0], 10, 64)
		if errParse != nil {
			return ProcDevices{}, fmt.Errorf("err parsing %s row:%s", errParse, row)
		}

		if readBlock {
			result.BlockDevices[major] = fie[1]
		}
		if readChar {
			result.CharacterDevices[major] = fie[1]
		}
	}
	return result, nil
}
