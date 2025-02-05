/*
stat demo
*/
package main

import (
	"fmt"
	"status"
)

func main() {
	sta, errSta := status.GetProcStat()
	if errSta != nil {
		fmt.Printf("ERR: %s\n", errSta)
		return
	}
	fmt.Printf("%#v\n", sta)
	fmt.Printf("\n------\n")
	for i, a := range sta.CPUs {
		fmt.Printf("%v:%.2f%%\n", i, a.CpuPercent())
	}
}
