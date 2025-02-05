package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"oledgui"

	"github.com/hjkoskel/gomonochromebitmap"
	"github.com/hjkoskel/listserialports"
	"github.com/hjkoskel/sds011"
	"github.com/hjkoskel/timegopher"
)

func ListSerialDevices() error {
	fmt.Printf("------ SERIAL PORT DETECT TEST ----------\n")
	drivers, driversListErr := listserialports.ListOfSerialTTYDriverPrefixes()
	if driversListErr != nil {
		return fmt.Errorf("Error listing serial drivers: %s", driversListErr)
	}
	fmt.Printf("Supported serial drivers %#v\n", drivers)

	entries, errProbe := listserialports.Probe(true)
	if errProbe != nil {
		fmt.Printf("%v\n", errProbe)
		os.Exit(-1)
	}
	for i, entry := range entries {
		fmt.Printf("%v: %v", i, entry.ToPrintoutFormat())
	}
	fmt.Printf("\n----------------------\n")
	return nil
}

const SAMPLEWORKBUFLEN = 60 * 24 * 30

func SetForDisplay(pic *gomonochromebitmap.MonoBitmap) error {
	picrender := gomonochromebitmap.BlockGraphics{
		Clear:       false,
		HaveBorder:  true,
		BorderColor: gomonochromebitmap.FGANSI_YELLOW,
		TextColor:   gomonochromebitmap.FGANSI_BLUE,
	}

	fmt.Print(picrender.ToQuadBlockChars(pic))
	fmt.Printf("\n\n")
	return nil
}

func MainRun(serialDeviceName string) error {
	resultCh := make(chan sds011.Result, 5)

	go func() {
		var workbuf ParticleMeasArr
		workbuf = make([]ParticleMeas, SAMPLEWORKBUFLEN)
		utCheck, _ := timegopher.CreateUptimeChecker()
		for {
			//fmt.Printf("\n\n-----------------\nwaiting resultch %v/%v\n", len(resultCh), cap(resultCh))
			m := <-resultCh

			tNow := time.Now()
			ut, _ := utCheck.UptimeNano(tNow)

			workbuf.Insert(ParticleMeas{
				BootNumber: 0,
				Uptime:     int64(ut),
				Epoch:      tNow.UnixMicro(),

				//Fill up latest information on ambient also
				Temperature: math.NaN(),
				Humidity:    math.NaN(),
				Pressure:    math.NaN(), //TODO SENSOR!
				Small:       m.Small(),
				Large:       m.Large(),
			}, SAMPLEWORKBUFLEN)

			pic := oledgui.ViewMetrics(m.Small(), m.Large(), false, math.NaN(), math.NaN(), math.NaN())
			SetForDisplay(&pic)
			//fmt.Printf("%v %v\n", tNow, m.ToString())
		}
	}()

	return RunSDS011(serialDeviceName, resultCh)
}

func RunSDS011(serialDeviceName string, resultch chan sds011.Result) error {
	serialLink, serialInitErr := sds011.CreateLinuxSerial(serialDeviceName)
	if serialInitErr != nil {
		return fmt.Errorf("Initializing serial port %v failed %v\n", serialDeviceName, serialInitErr.Error())
	}

	sensor := sds011.InitSds011(uint16(0xFFFF), false, serialLink, resultch, 0)
	defer serialLink.Close()
	return sensor.Run()

}
