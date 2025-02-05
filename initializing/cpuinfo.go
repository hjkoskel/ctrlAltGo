/*
CPU-info is good to know information when initialzing system
Differen behaviour on different systems
*/
package initializing

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type CpuCoreInfo map[string]string

func (a CpuCoreInfo) String() string {
	//TODO sort by alphabet? Makes stable and testable?
	keys := []string{}
	for k, _ := range a {
		keys = append(keys, k)
	}
	slices.Sort(keys) //sort by alphabet. Makes stable and testable
	var result strings.Builder
	for _, key := range keys {
		result.WriteString(fmt.Sprintf("%s : %s\n", key, a[key]))
	}
	return result.String()
}

func (p *CpuCoreInfo) Diff(ref CpuCoreInfo) CpuCoreInfo {
	result := make(map[string]string)
	for key, value := range *p {
		valueRef, haz := ref[key]
		if !haz || value != valueRef {
			result[key] = value //Missing from ref
		}
	}

	for refKey, refValue := range ref {
		v, haz := (*p)[refKey]
		_, resultHaveIt := result[refKey]
		if (!haz || refValue != v) && !resultHaveIt {
			result[refKey] = refValue
		}
	}

	return result
}

type CpuInfo []CpuCoreInfo

func (a CpuInfo) String() string {
	arr := make([]string, len(a))
	for i, a := range a {
		arr[i] = a.String()
	}
	return strings.Join(arr, "\n")
}

func (p *CpuInfo) Diff(ref CpuCoreInfo) CpuInfo {
	result := make([]CpuCoreInfo, len(*p))

	for i, c := range *p {
		result[i] = c.Diff(ref)
	}
	return result
}

func GetCpuinfo() (CpuInfo, error) {
	content, errRead := os.ReadFile("/proc/cpuinfo")
	if errRead != nil {
		return nil, errRead
	}

	s := strings.ReplaceAll(string(content), "\t", "")
	coreContents := strings.Split(s, "\n\n")
	result := []CpuCoreInfo{}

	//fmt.Printf("coreContents %#v\n", coreContents)

	for _, cor := range coreContents {
		if len(strings.TrimSpace(cor)) == 0 {
			continue
		}
		rows := strings.Split(cor, "\n")
		m := make(map[string]string)
		for _, row := range rows {
			cols := strings.Split(row, ":")
			if len(cols) < 2 {
				return result, fmt.Errorf("unexpected row with no separator %s", row)
			}
			// strings.TrimSpace(cols[1])
			m[strings.TrimSpace(cols[0])] = strings.TrimSpace(strings.Join(cols[1:], ":"))
		}
		result = append(result, m)
	}

	return result, nil

}
