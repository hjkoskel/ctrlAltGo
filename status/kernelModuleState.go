package status

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// KernelModule represents a single kernel module entry from /proc/modules
type KernelModuleState struct {
	Name       string       //0
	Size       int64        //1
	Instances  int64        //2
	UsedBy     []string     //3
	State      ModuleStatus //4
	Address    uint64       //5
	Annotation string       //6,optional
}

func ParseKernelModuleState(line string) (KernelModuleState, error) {
	var result KernelModuleState

	fie := strings.Split(line, " ")
	if len(fie) < 6 {
		return result, fmt.Errorf("INVALID LINE:\n%s\n", line)
	}
	fmt.Printf("Fields %#v\n", fie)

	result.Name = fie[0]

	var errParse error
	result.Size, errParse = strconv.ParseInt(fie[1], 10, 64)
	if errParse != nil {
		return result, fmt.Errorf("parse error fields=%#v err:%s", fie, errParse)
	}

	result.Instances, errParse = strconv.ParseInt(fie[2], 10, 64)
	if errParse != nil {
		return result, fmt.Errorf("parse error fields=%#v err: %s", fie, errParse)
	}

	result.UsedBy = []string{}
	usedlist := strings.Split(fie[3], ",")
	for _, name := range usedlist {
		if name == "-" || len(name) == 0 {
			continue
		}
		result.UsedBy = append(result.UsedBy, name)
	}

	switch fie[4] {
	case LIVE, LOADING, UNLOADING:
		result.State = ModuleStatus(fie[4])
	default:
		return result, fmt.Errorf("unknow status %s", fie[4])
	}

	result.Address, errParse = strconv.ParseUint(strings.Replace(fie[5], "0x", "", 1), 16, 64)
	if errParse != nil {
		return result, fmt.Errorf("parse error fields=%#v err:%s", fie, errParse)
	}
	result.Annotation = ""
	if 6 < len(fie) {
		result.Annotation = strings.Replace(fie[6], "(", "", 1)
		result.Annotation = strings.Replace(result.Annotation, "(", "", 1)
	}

	return result, nil
}

type ModuleStatus string

const (
	LIVE      = "Live"
	LOADING   = "Loading"
	UNLOADING = "Unloading"
)

/*
kfifo_buf 12288 0 - Live 0xffffffe126ab2000
industrialio 98304 1 kfifo_buf, Live 0xffffffe126a97000
*/

func ReadKernelModuleStates(procModulesFile string) ([]KernelModuleState, error) {
	// Read the entire /proc/modules file
	data, err := os.ReadFile(procModulesFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s err:%s", procModulesFile, err)
	}
	lines := strings.Split(string(data), "\n")
	result := []KernelModuleState{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parsed, errParse := ParseKernelModuleState(line)
		if errParse != nil {
			return result, errParse
		}
		result = append(result, parsed)
		fmt.Printf("%s\n%#v\n\n", line, parsed)
	}
	return result, nil
}
