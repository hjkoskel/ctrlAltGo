/*
Simple SDS010 particle meter

- 128x64 display
- disk partition for storing settings and log data?
- ethernet (or WLAN connectivity)

*/

package main

import (
	"fmt"
	"os"
	"time"

	_ "time/tzdata" //Smart thing to have up to date tzdata here

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/hjkoskel/ctrlaltgo/networking"

	"github.com/hjkoskel/timegopher"
	"github.com/hjkoskel/timegopher/timesync"
)

const INTERFACENAME = "eth0"
const HOSTNAMEOFSYSTEM = "simplesrv"
const TIMEZONENAME = "Europe/Helsinki" //TODO CONFIG?

var tz *time.Location

func initialInit() error {
	var err error
	tz, err = time.LoadLocation(TIMEZONENAME)
	if err != nil {
		return fmt.Errorf("tz err %s", err)
	}
	err = initializing.MountNormal()
	if err != nil {
		return fmt.Errorf("error mounting normal %s", err)
	}

	err = initializing.SetHostname(HOSTNAMEOFSYSTEM) //Important when having network
	if err != nil {
		return err
	}

	/*
		hostnameByKernel, errHostname := os.Hostname()
		ctrlaltgo.JamIfErr(errHostname)

		fmt.Printf("hostname by kernel:%s\n", hostnameByKernel)

		listDir("/")
	*/

	//Doing network setup

	fmt.Printf("Bring %v up\n", INTERFACENAME)

	errWaitInterf := networking.WaitInterface(INTERFACENAME, time.Second*30, time.Second) //Raspberry pi delay?
	if errWaitInterf != nil {
		return fmt.Errorf("error waiting %s interface err:%s", INTERFACENAME, errWaitInterf)
	}

	errUp := networking.SetLinkUp(INTERFACENAME, true)
	if errUp != nil {
		return fmt.Errorf("error bringing up interface %s err:%s", INTERFACENAME, errUp)
	}
	return nil
}

// This routine tries to keep network up TODO Channel for network settings change!
func keepNetworkUp() error {
	//THinking... For creating network there is only inteface working on this device wlan or LAN
	fmt.Printf("Going to keep net up!\n")
	for {
		haveCarr, errCarr := networking.Carrier(INTERFACENAME)
		if errCarr != nil {
			return fmt.Errorf("checking carrier err:%s", errCarr)
		}
		if haveCarr {
			break
		}
		fmt.Printf(".... no carrier\n")
	}
	fmt.Printf("... have carrier\n")

	//Run in goroutine?

	ipSettings, errDhcp := networking.GetDHCP(HOSTNAMEOFSYSTEM, INTERFACENAME)
	if errDhcp != nil {
		return errDhcp
	}

	fmt.Printf("GOT IP settings %#v\n", ipSettings)

	errApplyIp := ipSettings.ApplyToInterface(INTERFACENAME, 1)
	if errApplyIp != nil {
		return errApplyIp
	}

	for time.Now().Before(ipSettings.Expire) {
		time.Sleep(time.Millisecond * 1500)
		haveCarr, errCarr := networking.Carrier(INTERFACENAME)
		if errCarr != nil {
			return fmt.Errorf("checking carrier err:%s", errCarr)
		}
		if !haveCarr {
			return fmt.Errorf("lost carrier")
		}
	}
	return fmt.Errorf("ip lease expired after %s (at %s)", ipSettings.LeaseTime, ipSettings.Expire)
}

func clockSyncRoutine(checkInterval time.Duration) error {
	for {
		fmt.Printf("TIME before sync IS NOW %s\n", time.Now().In(tz))
		var timeOffset time.Duration
		var err error
		for {
			ntpsync := timesync.GetDefaultFinnishNTP()
			timeOffset, err = ntpsync.GetOffset()
			if err == nil {
				break
			}
			time.Sleep(time.Second * 5)
			fmt.Printf("getting ntp err:%s", err)
		}

		fmt.Printf("Got timeoffet %s\n", timeOffset)
		errSet := timegopher.SetSysClock(time.Now().Add(timeOffset))
		if errSet != nil {
			return errSet
		}
		time.Sleep(checkInterval)
	}
}

func main() {
	if os.Getpid() != 1 {
		fmt.Printf("-- Running particle meter under OS --\n")
		errRun := MainRun("/dev/ttyS0")
		if errRun != nil {
			fmt.Printf("\nMAIN ERROR:%s\n", errRun)
			return
		}
		fmt.Printf("Software failed with no error\n")
		return
	}

	fmt.Printf("Initial initalization....\n")
	errInitInitial := initialInit()
	ctrlaltgo.JamIfErr(errInitInitial)

	/*
		go func() {
			for {
				errNetwork := keepNetworkUp()
				fmt.Printf("TODO STATUS UPDATE NETWORK ERR %s\n", errNetwork)
				time.Sleep(time.Second * 3)
			}
		}()

		go func() {
			for {
				errClkSync := clockSyncRoutine(time.Hour)
				fmt.Printf("CLOCK SYNC FAIL! %s\n", errClkSync)
			}
		}()
	*/

	errRun := MainRun("/dev/ttyS0")
	if errRun != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("\nMAIN ERROR:%s\n", errRun))

	}
	ctrlaltgo.JamIfErr(fmt.Errorf("Software failed with no error\n"))

}
