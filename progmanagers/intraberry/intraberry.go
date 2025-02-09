/*
Very simple web browser based development program manager

Usable only intranet. Non cybersafe.
- read config files from sdcard
- webui
  - Upload program
  - Upload initramfs
  - config.txt management?

- some misc status

- Usable from web browser. Leave console for actual application
*/
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	_ "time/tzdata" //Smart thing to have up to date tzdata here

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/hjkoskel/ctrlaltgo/networking"
	"github.com/hjkoskel/timegopher"
)

const (
	DEV_BOOTCARD  = "/dev/mmcblk0p1"
	MNT_BOOTCARD  = "/bootcard"
	PROGRAM       = "/bootcard/program"
	TMPPROGRAM    = "/tmp/program"
	SETTINGSDIR   = MNT_BOOTCARD + "/intraberry"
	INTERFACENAME = "eth0"

	PAUSEPROG = "STOP" //If program name is this, stops current and wait next. Prevents having binary in use while updating
)

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

var onlineNow bool

func mainNetwork(hostname string, ipSettings *networking.IpSettings) error {
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

	if ipSettings != nil {
		errSet := ipSettings.ApplyToInterface(INTERFACENAME, 1)
		if errSet != nil {
			return fmt.Errorf("error settting fixed IP settings %#v err:%s", ipSettings, errSet)
		}
		return nil
	}

	//Run in goroutine?
	//If clock is wrong

	for {
		ipSettingsDhcp, errDhcp := networking.GetDHCP(hostname, INTERFACENAME)
		if errDhcp != nil {
			fmt.Printf("DHCP err %s\n", errDhcp)
			time.Sleep(time.Second * 3)
			continue
		}
		tGotIP := time.Now()
		fmt.Printf("GOT DHCP settings %s\n", ipSettingsDhcp)
		errApplyIp := ipSettingsDhcp.ApplyToInterface(INTERFACENAME, 1)
		if errApplyIp != nil {
			fmt.Printf("error settting DHCP IP settings %#v err:%s", ipSettings, errApplyIp)
			time.Sleep(time.Second * 3)
			continue
		}
		fmt.Printf("settings applyed\n")
		time.Sleep(ipSettingsDhcp.LeaseTime - time.Since(tGotIP))
	}
}

// ExecOneProgram executes one program at a time.
// When a new program name is received on the programQueue channel, it stops the current program and starts the new one.
// Program output is sent to normalOut, and errors are sent to errPrintout.
// Returns name of program that was running and an error if the program crashes or if there's an issue with the queue.
func ExecOneProgram(programQueue chan string, normalOut chan string, errPrintout chan string, errorsCh chan error) {
	var cmd *exec.Cmd
	var cancel context.CancelFunc
	var wg sync.WaitGroup

	progNow := PROGRAM
	//TODO external monitoring of channel? Keep in crashed state? and wait re-try?
	for program := range programQueue {
		errorsCh <- fmt.Errorf("----")
		if len(program) == 0 {
			program = progNow //Allows restarting previous with empty string
		}

		//progNow = program
		// If a program is already running, stop it
		if cancel != nil {
			cancel()
			wg.Wait() // Wait for the current program to exit
		}

		for program == PAUSEPROG { //wait pause
			fmt.Printf("ON PAUSE\n")
			program = <-programQueue
		}

		fmt.Printf("\n\n*** EXECUTING: %s ***\n", program)

		// Create a new context for the new program
		ctx, ctxCancel := context.WithCancel(context.Background())
		cancel = ctxCancel

		// Start the new program
		cmd = exec.CommandContext(ctx, program)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errorsCh <- fmt.Errorf("error creating stdout pipe for %s: %v", program, err)
			continue
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			errorsCh <- fmt.Errorf("error creating stderr pipe for %s: %v", program, err)
			continue
		}

		// Start the program
		if err := cmd.Start(); err != nil {
			errorsCh <- fmt.Errorf("error starting program %s: %v", program, err)
			continue
		}

		// Read stdout and stderr in goroutines
		wg.Add(2)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				normalOut <- scanner.Text()
			}
		}()
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				errPrintout <- scanner.Text()
			}
		}()

		// Wait for the program to finish in a goroutine
		go func() {
			err := cmd.Wait()
			if err != nil {
				errorsCh <- fmt.Errorf("program %s crashed: %v", program, err)
			} else {
				errorsCh <- fmt.Errorf("program %s exited successfully", program)
			}
		}()
	}

	// Clean up if the programQueue channel is closed
	if cancel != nil {
		cancel()
		wg.Wait()
	}
}

func execApp(ctx context.Context, programFileName string, normalOut chan string, errPrintout chan string) error {
	if !FileExists(programFileName) {
		return fmt.Errorf("program file %s not found", programFileName)
	}

	cmd := exec.CommandContext(ctx, programFileName)
	if cmd == nil {
		return fmt.Errorf("exec.Command returns nil for program")
	}

	stdReader, errStdReader := cmd.StdoutPipe()
	if errStdReader != nil {
		return fmt.Errorf("stdout pipe err %s", errStdReader)
	}
	errReader, errErrReader := cmd.StderrPipe()
	if errErrReader != nil {
		return fmt.Errorf("stderr pipe err %s", errErrReader)
	}

	scanStd := bufio.NewScanner(stdReader)
	scanErr := bufio.NewScanner(errReader)
	go func() {
		for scanStd.Scan() {
			normalOut <- scanStd.Text()
		}
	}()

	go func() {
		for scanErr.Scan() {
			errPrintout <- scanErr.Text()
		}
	}()

	errStart := cmd.Start()
	if errStart != nil {
		return fmt.Errorf("error starting %s", errStart)
	}

	errWait := cmd.Wait()
	if errWait != nil {
		return fmt.Errorf("error while waiting program %s", errWait)
	}

	return fmt.Errorf("program exitted wait nil err")
}

func main() {
	if os.Getpid() != 1 {
		fmt.Printf("This is initramfs program. Please run this as PID1\n")
		return
	}
	fmt.Printf("\n\n\n-- Intraberry --\n\n\n")

	fmt.Printf("mounting normal set of dirs\n")
	errMount := initializing.MountNormal()
	ctrlaltgo.JamIfErr(errMount)

	initializing.CreateAndMount([]initializing.MountCmd{
		initializing.MountCmd{
			Source: DEV_BOOTCARD,
			Target: MNT_BOOTCARD,
			FsType: "vfat",
			//Flags  uintptr
			//Data   string
		},
	})

	hostname, _ := GetHostname(SETTINGSDIR) //Skip errors? run with default. TODO where report?
	fmt.Printf("Hostname is %s\n", hostname)
	errSetHostname := initializing.SetHostname(hostname)
	ctrlaltgo.JamIfErr(errSetHostname)
	tz, tzGetErr := GetTz(SETTINGSDIR)

	ctrlaltgo.JamIfErr(tzGetErr)

	hostnameByKernel, errHostname := os.Hostname()
	ctrlaltgo.JamIfErr(errHostname)

	fmt.Printf("hostname by kernel:%s\n", hostnameByKernel)

	go func() {
		fmt.Printf("running on tz=%s", tz.String())

		fmt.Printf("TIME before sync IS NOW %s\n", time.Now().In(tz))
		ntpsync, _ := GetNtpSettings(SETTINGSDIR)

		//time.Sleep(time.Second * 5000) //Catch kernel panic

		for {
			timeOffset, errNtp := ntpsync.GetOffset()
			if errNtp != nil {
				time.Sleep(time.Minute * 5)
				continue
			}
			fmt.Printf("Got timeoffet %s\n", timeOffset)
			errSetClock := timegopher.SetSysClock(time.Now().Add(timeOffset))
			ctrlaltgo.JamIfErr(errSetClock)
			fmt.Printf("TIME after sync IS NOW %s\n", time.Now().In(tz))
			time.Sleep(time.Hour)
		}
	}()
	//

	go func() {

		manualEthSettings, _ := GetEthSettings(SETTINGSDIR)
		networkFail := mainNetwork(hostname, manualEthSettings)
		ctrlaltgo.JamIfErr(networkFail)
	}()

	//MNT_BOOTCARD
	fmt.Printf("---Settings dir have have---\n")
	listDir(path.Dir(SETTINGSDIR))

	fmt.Printf("---Program dir have have---\n")
	listDir(path.Dir(PROGRAM))

	normalOut := make(chan string, 1000)
	errPrintout := make(chan string, 1000)

	progCrashQueue := make(chan error, 100) //Should not happen but if happens! Let UI know!
	progExecQueue := make(chan string, 1)
	progExecQueue <- PROGRAM //Initial value
	go ExecOneProgram(progExecQueue, normalOut, errPrintout, progCrashQueue)

	//Start server
	errRunServer := MaintananceServer(progExecQueue, normalOut, errPrintout, progCrashQueue)
	ctrlaltgo.JamIfErr(errRunServer)

	ctrlaltgo.JamIfErr(fmt.Errorf("-- DONE --"))
}
