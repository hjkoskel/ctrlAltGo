/*
Example program to demonstrate how to set up network and run webserver there
*/
package main

import (
	"ctrlaltgo"
	"fmt"
	"initializing"
	"io"
	"net/http"
	"networking"
	"os"
	"time"
	_ "time/tzdata" //Smart thing to have up to date tzdata here

	"github.com/hjkoskel/timegopher"
	"github.com/hjkoskel/timegopher/timesync"
)

const INTERFACENAME = "eth0"
const HOSTNAMEOFSYSTEM = "simplesrv"
const TIMEZONENAME = "Europe/Helsinki"

/*
https://www.qemu.org/docs/master/system/qemu-manpage.html
-rtc

TODO test Clock sync
*/

func doNetworkTest() {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("http://www.example.com") //"http://192.168.1.106:2222")
	if err != nil {
		fmt.Printf("Connectivity test failed %s better luck next time\n", err.Error())
	} else {
		fmt.Printf("CONNECTED!\n")
		body, errBody := io.ReadAll(resp.Body)
		if errBody != nil {
			fmt.Printf("ERROR loading body %s\n\n", errBody)
		} else {
			fmt.Printf("%s\n\n", body)
		}
	}
}

func listDir(dirname string) error {
	entries, errRead := os.ReadDir(dirname)
	if errRead != nil {
		return errRead
	}

	if errRead != nil {
		return errRead
	}

	fmt.Printf("DIR: %s  ", dirname)
	for _, entry := range entries {
		fmt.Printf("%s", entry.Name())
		if entry.IsDir() {
			fmt.Printf("(D)  ")
		} else {
			fmt.Printf("(F)  ")
		}
	}
	fmt.Printf("\n\n")
	return nil
}

func main() {
	if os.Getpid() != 1 {
		fmt.Printf("This is initramfs program. Please run this as PID1\n")
		return
	}
	fmt.Printf("-- Starting example init program --\n")

	tz, errTz := time.LoadLocation("Europe/Helsinki")
	ctrlaltgo.JamIfErr(errTz)

	fmt.Printf("mounting normal set of dirs\n")
	errMount := initializing.MountNormal()
	ctrlaltgo.JamIfErr(errMount)

	errSetHostname := initializing.SetHostname(HOSTNAMEOFSYSTEM) //Important when having network
	ctrlaltgo.JamIfErr(errSetHostname)

	hostnameByKernel, errHostname := os.Hostname()
	ctrlaltgo.JamIfErr(errHostname)

	fmt.Printf("hostname by kernel:%s\n", hostnameByKernel)

	listDir("/")

	//Doing network setup

	fmt.Printf("Bring %v up\n", INTERFACENAME)

	errWaitInterf := networking.WaitInterface(INTERFACENAME, time.Second*30, time.Second) //Raspberry pi delay?
	ctrlaltgo.JamIfErr(errWaitInterf)

	errUp := networking.SetLinkUp(INTERFACENAME, true)
	ctrlaltgo.JamIfErr(errUp)
	for {
		fmt.Printf("Checking carrier on %s ....", INTERFACENAME)
		haveCarr, errCarr := networking.Carrier(INTERFACENAME)
		ctrlaltgo.JamIfErr(errCarr)
		if haveCarr {
			break
		}
		fmt.Printf(".... no carrier\n")
	}
	fmt.Printf("... have carrier\n")

	//Run in goroutine?

	ipSettings, errDhcp := networking.GetDHCP(HOSTNAMEOFSYSTEM, INTERFACENAME)
	ctrlaltgo.JamIfErr(errDhcp)

	fmt.Printf("GOT IP settings %#v\n", ipSettings)

	errApplyIp := ipSettings.ApplyToInterface(INTERFACENAME, 1)
	ctrlaltgo.JamIfErr(errApplyIp)

	doNetworkTest()

	fmt.Printf("TIME before sync IS NOW %s\n", time.Now().In(tz))

	ntpsync := timesync.GetDefaultFinnishNTP()

	timeOffset, errNtp := ntpsync.GetOffset()
	ctrlaltgo.JamIfErr(errNtp)
	fmt.Printf("Got timeoffet %s\n", timeOffset)
	errSetClock := timegopher.SetSysClock(time.Now().Add(timeOffset))
	ctrlaltgo.JamIfErr(errSetClock)
	fmt.Printf("TIME after sync IS NOW %s\n", time.Now().In(tz))
	//

	//Start server
	errRunServer := RunExampleWebServer()
	ctrlaltgo.JamIfErr(errRunServer)

	ctrlaltgo.JamIfErr(fmt.Errorf("-- DONE --"))
}
