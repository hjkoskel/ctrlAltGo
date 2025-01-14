/*
Run as root and see keypresses
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
	keyboards := devArr.GetDefaultKeyboards()

	pKeyboardNumber := flag.Int("i", 0, fmt.Sprintf("index of keyboard, got %v", len(keyboards)))
	flag.Parse()

	if *pKeyboardNumber < 0 || len(keyboards) <= *pKeyboardNumber {
		fmt.Printf("invalid keyboard index, got only %v keyboards\n", len(keyboards))
		return
	}

	kbFileName := keyboards[*pKeyboardNumber].GetDevName()

	fDev, errFDev := os.Open(kbFileName)
	if errFDev != nil {
		fmt.Printf("error reading device %s  err:%s\n", kbFileName, errFDev)
		return
	}
	keyboard := inputdev.InitKeyboard()
	for {
		ev, errEv := inputdev.ReadInputEvent(fDev)
		if errEv != nil {
			fmt.Printf("error reading event %s\n", errEv)
			return
		}
		keyboard.Update(&ev)
		fmt.Printf("%#v\n", keyboard.Buffer)
		fmt.Printf("TEXT: %s\n", keyboard.Text)
		fmt.Printf("PRESSED: %#v\n", keyboard.PressedKeys())
		if 0 < keyboard.State[inputdev.KEY_ENTER] {
			keyboard.Buffer = []inputdev.KeyCode{inputdev.KEY_ENTER}
			keyboard.Text = ""
		}
	}

}
