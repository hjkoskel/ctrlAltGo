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

type CpuInfo struct {
	Common CpuCoreInfo
	Cores  []CpuCoreInfo
}

func (a CpuInfo) String() string {
	var sb strings.Builder

	for i, a := range a.Cores {
		sb.WriteString(fmt.Sprintf("-- Core%d --\n%s\n\n", i, a.String()))
	}
	sb.WriteString(fmt.Sprintf("-- Common ---\n%s\n\n", a.Common))

	return sb.String()
}

func (p *CpuCoreInfo) GetCommon(other CpuCoreInfo) CpuCoreInfo {
	result := make(map[string]string)
	for key, value := range *p {
		a, hazA := (*p)[key]
		b, hazB := other[key]

		if hazA && hazB && a == b {
			result[key] = value
		}
	}
	return result
}

func (p *CpuInfo) Commonize() {
	if len(p.Cores) < 2 {
		return
	}
	com := p.Cores[0].GetCommon(p.Cores[1])
	for _, c := range p.Cores {
		com = c.GetCommon(com)
	}
	if p.Common == nil {
		p.Common = CpuCoreInfo{}
	}
	for key, value := range com {
		p.Common[key] = value
		for i := range p.Cores {
			delete(p.Cores[i], key)
		}
	}
}

/*
func (p *CpuInfo) Diff() CpuInfo {
	result := CpuInfo{Common: p.Common}
	if len(p.Cores) == 0 {
		return CpuInfo{}
	}
	if result.Common == nil {
		result.Common = CpuCoreInfo{} //make(map[string]string)
	}
	for key, value := range p.Cores[0] {
		result.Common[key] = value
	}

	result.Cores = make([]CpuCoreInfo, len(p.Cores))
	for i, c := range p.Cores {
		result.Cores[i] = c.Diff(p.Cores[0])
	}
	return result
}*/

func GetCpuinfo(cpuinfofilename string) (CpuInfo, error) {
	content, errRead := os.ReadFile(cpuinfofilename)
	if errRead != nil {
		return CpuInfo{}, errRead
	}

	s := strings.ReplaceAll(string(content), "\t", "")
	coreContents := strings.Split(s, "\n\n")
	result := CpuInfo{}

	for _, cor := range coreContents {
		if len(strings.TrimSpace(cor)) == 0 {
			continue
		}
		rows := strings.Split(strings.TrimSpace(cor), "\n")
		m := make(map[string]string)
		firstkey := ""
		for rowcounter, row := range rows {
			cols := strings.Split(row, ":")
			if len(cols) < 2 {
				return result, fmt.Errorf("unexpected row with no separator %s", row)
			}
			key := strings.TrimSpace(cols[0])
			if rowcounter == 0 {
				firstkey = key
			}
			// strings.TrimSpace(cols[1])
			m[key] = strings.TrimSpace(strings.Join(cols[1:], ":"))
		}
		if firstkey == "processor" {
			result.Cores = append(result.Cores, m)
		} else {
			if 0 < len(result.Common) {
				return result, fmt.Errorf("unexpected data format more than 1 common part on /proc/cpuinfo")
			}
			result.Common = m
		}
	}
	//Core without processor number is common part
	return result, nil
}
