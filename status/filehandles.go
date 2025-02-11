package status

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type FileHandlesUsage struct {
	Allocated int64
	Free      int64 //Is 0 on modern systems
	Maximum   int64
}

func (a FileHandlesUsage) String() string {
	if a.Maximum == 0 {
		return "invalid maximum=0"
	}
	return fmt.Sprintf("%v/%v = %.2f%%", a.Allocated, a.Maximum, float64(10000*a.Allocated/a.Maximum)/100)
}

func GetFileHandlesUsage() (FileHandlesUsage, error) {
	byt, errRead := os.ReadFile("/proc/sys/fs/file-nr")
	if errRead != nil {
		return FileHandlesUsage{}, errRead
	}
	s := strings.TrimSpace(string(byt))
	cols := strings.Fields(s)
	if len(cols) != 3 {
		return FileHandlesUsage{}, fmt.Errorf("invalid file handles file:%s", s)
	}
	var result FileHandlesUsage
	var errConv error
	result.Allocated, errConv = strconv.ParseInt(cols[0], 10, 64)
	if errConv != nil {
		return FileHandlesUsage{}, fmt.Errorf("invalid file handles file:%s  err:%w", s, errConv)
	}
	result.Free, errConv = strconv.ParseInt(cols[1], 10, 64)
	if errConv != nil {
		return FileHandlesUsage{}, fmt.Errorf("invalid file handles file:%s  err:%w", s, errConv)
	}
	result.Maximum, errConv = strconv.ParseInt(cols[2], 10, 64)
	if errConv != nil {
		return FileHandlesUsage{}, fmt.Errorf("invalid file handles file:%s  err:%w", s, errConv)
	}
	return result, nil
}
