package main

import (
	"fmt"
	"time"

	"github.com/hjkoskel/govattu"
)

//const LEDPIN = 21 //or 40?

const LED = 21

var txtOnTime string = "1s" //-ldflags works only strings
var txtOffTime string = "2s"

func infiniteBlink(onDur time.Duration, offDur time.Duration) {
	fmt.Printf("ON:%s  OFF:%s\n", onDur, offDur)
	hw, errOpen := govattu.Open()
	if errOpen != nil {
		fmt.Printf("error open  %s\n", errOpen)
		return
	}
	hw.PinMode(LED, 1)
	for {
		time.Sleep(onDur)
		hw.PinSet(LED)
		time.Sleep(offDur)
		hw.PinClear(LED)
	}
}

func main() {
	toff, errToff := time.ParseDuration(txtOffTime)
	if errToff != nil {
		fmt.Printf("INTERNAL ERR txtOffTime=%s  err:%s\n", txtOffTime, errToff)
		return
	}
	ton, errTon := time.ParseDuration(txtOnTime)
	if errTon != nil {
		fmt.Printf("INTERNAL ERR txtOnTime=%s  err:%s\n", txtOnTime, errTon)
		return
	}

	infiniteBlink(ton, toff)
}
