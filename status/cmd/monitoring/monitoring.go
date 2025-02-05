/*
Minimal example
*/
package main

import (
	"fmt"
	"time"

	"status"
)

func main() {
	msgCh := make(chan status.KMsg, 100000)
	go func() {
		for {
			m := <-msgCh
			fmt.Printf("%s\n", m.String())
		}
	}()

	mon, errOpen := status.OpenKernelMonitor(1024 * 16)
	if errOpen != nil {
		fmt.Printf("open err %s\n", errOpen)
		return
	}

	for {
		errRead := mon.Read(msgCh)
		if errRead != nil {
			fmt.Printf("READING failed:%s\n", errRead)
			return
		}
		fmt.Printf("----- read failed with no reason----\n")
		time.Sleep(time.Second * 2)
	}
}
