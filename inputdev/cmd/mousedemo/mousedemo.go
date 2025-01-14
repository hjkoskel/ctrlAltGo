/*
Run as root and see events of mouse
TODO graphical?
*/
package main

import (
	"flag"
	"fmt"
	"inputdev"
	"os"
)

func main() {

	devArr, errParse := inputdev.ParseDevicesFile(inputdev.INPUTDEVICESFILE)
	if errParse != nil {
		fmt.Printf("err:%s\n", errParse)
	}
	mices := devArr.GetMices()

	pMouseNumber := flag.Int("i", 0, fmt.Sprintf("index of mouse, got %v", len(mices)))
	flag.Parse()

	if *pMouseNumber < 0 || len(mices) <= *pMouseNumber {
		fmt.Printf("invalid keyboard index, got only %v keyboards\n", len(mices))
		return
	}

	mouseFileName := mices[*pMouseNumber].GetDevName()

	fDev, errFDev := os.Open(mouseFileName)
	if errFDev != nil {
		fmt.Printf("error reading device %s  err:%s\n", mouseFileName, errFDev)
		return
	}

	//TODO proper init when it is decided how to track position of cursor
	mouse := inputdev.Mouse{}
	for {
		ev, errEv := inputdev.ReadInputEvent(fDev)
		if errEv != nil {
			fmt.Printf("error reading event %s\n", errEv)
			return
		}
		mouse.Update(&ev)

		fmt.Printf("EV: %#v\n", ev)
		fmt.Printf("Clicks: %#v rel:%#v\n", mouse.Clicks, mouse.RelCoord)

	}

}
