package status

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MemInfo map[string]int64 //In bytes

const (
	MEMINFOFILE = "/proc/meminfo"
)

func ReadMemInfo(memFinfoFileName string) (MemInfo, error) {
	byt, errRead := os.ReadFile(memFinfoFileName)
	if errRead != nil {
		return MemInfo{}, errRead
	}
	rows := strings.Split(string(byt), "\n")

	result := make(map[string]int64)
	for _, row := range rows {

		fie := strings.Fields(row)
		if len(fie) < 2 {
			continue
		}

		if len(fie) == 3 {
			if fie[2] != "kB" {
				return result, fmt.Errorf("invalid row %s", row)
			}
		}

		varName := strings.TrimSpace(strings.Replace(fie[0], ":", "", 1))
		n, parseErr := strconv.ParseInt(fie[1], 10, 64)
		if parseErr != nil {
			return result, fmt.Errorf("error parsing field1 %#v  err:%s", fie, parseErr)
		}
		result[varName] = n * 1024
	}
	return result, nil
}
