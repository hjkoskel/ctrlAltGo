/*
Search set of drivers and prints out loading order of modules
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type DriverPath string
type DriverPaths []DriverPath

func (p *DriverPath) Name() string {
	s := path.Base(string(*p))
	s = strings.Replace(s, ".ko.xz", "", 1)
	return strings.Replace(s, ".ko", "", 1)
}

func (p *DriverPaths) Contains(d DriverPath) bool {
	for _, a := range *p {
		if a == d {
			return true
		}
	}
	return false
}

func (a DriverPaths) String() string {
	arr := make([]string, len(a))
	for i, s := range a {
		arr[i] = string(s)
	}
	return strings.Join(arr, "\n")
}

type DependencyFile map[DriverPath][]DriverPath

func RemoveRepeats(lst []DriverPath) DriverPaths {
	//Remove last ones
	var result DriverPaths
	for _, a := range lst {
		if !result.Contains(a) {
			result = append(result, a)
		}
	}
	return result
}

// List depend. First list all items without dependencies (loading order)
func (p *DependencyFile) ListDepend(d DriverPath) DriverPaths {
	result := []DriverPath{}
	deps, haz := (*p)[d]
	if !haz {
		return result
	}
	for _, dep := range deps {
		result = append(result, p.ListDepend(dep)...)
	}
	//Finalize
	result = append(result, d)

	return RemoveRepeats(result)
}

func (p *DependencyFile) GetDriverMatches(word string) []DriverPath { //Return multiples if not clear
	//Direct full path matches..ok
	direct, hazDirect := (*p)[DriverPath(word)]
	if hazDirect {
		return []DriverPath(direct)
	}
	//name mathes
	result := []DriverPath{}
	if strings.HasSuffix(word, ".ko.xz") || strings.HasSuffix(word, ".ko") || strings.HasSuffix(word, ".ko.zst") {
		for s := range *p {
			if word == path.Base(string(s)) {
				result = append(result, s)
			}
		}
		return result
	}
	//name without suffix
	for s := range *p {
		if word == s.Name() {
			result = append(result, s)
		}
	}
	if 0 < len(result) {
		return result //enough
	}
	//any matches
	for s := range *p {
		if strings.Contains(s.Name(), word) {
			result = append(result, s)
		}
	}
	return result
}

func ReadDepenencyFile(fname string) (DependencyFile, error) {
	byt, errRead := os.ReadFile(fname)
	if errRead != nil {
		return DependencyFile{}, errRead
	}
	rows := strings.Split(string(byt), "\n")
	result := make(map[DriverPath][]DriverPath)
	for i, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) == 0 {
			continue
		}
		cols := strings.Split(row, ":")
		if len(cols) != 2 {
			return result, fmt.Errorf("invalid row(%d) %s", i, row)
		}
		driver := strings.TrimSpace(cols[0])
		dependencies := strings.Split(cols[1], " ")

		_, haz := result[DriverPath(driver)]
		if haz {
			return result, fmt.Errorf("error duplicate driver %s on line %d", driver, i)
		}
		drvlist := make([]DriverPath, len(dependencies))
		for j, v := range dependencies {
			drvlist[j] = DriverPath(v)
		}
		result[DriverPath(driver)] = drvlist
	}
	return result, nil
}

func readOneLineFile(fname string) (string, error) {
	byt, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(byt)), nil
}

const OSRELEASEFILENAME = "/proc/sys/kernel/osrelease"

func main() {

	osrelease, errOsRelease := readOneLineFile(OSRELEASEFILENAME)
	if errOsRelease != nil {
		fmt.Printf("Internal error reading %s err:%s", OSRELEASEFILENAME, errOsRelease)
		return
	}
	//---- flags -------
	pDepFilename := flag.String("dep", fmt.Sprintf("/lib/modules/%s/modules.dep", osrelease), "module dependencies file")
	pNameList := flag.String("n", "", "comma separated list of driver names needed on installation")
	flag.Parse()
	//---------------
	namestring := strings.TrimSpace(*pNameList)
	if len(namestring) == 0 {
		fmt.Printf("No input\n")
		return
	}

	nameArr := strings.Split(namestring, ",")

	db, errDbLoad := ReadDepenencyFile(*pDepFilename)
	if errDbLoad != nil {
		fmt.Printf("Error loading dependencies err:%s\n", errDbLoad)
		return
	}

	//Search what are required
	driverList := make([]DriverPath, len(nameArr))
	for i, name := range nameArr {
		matches := db.GetDriverMatches(name)
		if len(matches) == 0 {
			fmt.Printf("No matches for %s\n", name)
		}
		if 1 < len(matches) {
			fmt.Printf("Multiple matches (%d) for %s  =>\n%s\n", len(matches), name, DriverPaths(matches).String())
		}
		driverList[i] = matches[0]
	}
	//Then start filling result array
	result := []DriverPath{}
	for _, drv := range driverList {
		result = append(result, db.ListDepend(drv)...)
	}

	result = RemoveRepeats(result)

	kerneldir := path.Dir(*pDepFilename)
	for i, v := range result {
		result[i] = DriverPath(path.Join(kerneldir, string(v)))
	}

	fmt.Printf("%s\n", DriverPaths(result))
}
