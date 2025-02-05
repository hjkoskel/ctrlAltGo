package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"os"
	"os/exec"
	"strings"
	"time"

	"textmonitor"

	"inputdev"

	"status"

	"initializing"
	"networking"

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/gomonochromebitmap"
	"golang.org/x/term"
)

func CreateDemoPic() gomonochromebitmap.MonoBitmap {
	result := gomonochromebitmap.NewMonoBitmap(128, 64, false)
	testfont := gomonochromebitmap.GetFont_8x8()
	result.Circle(image.Point{X: 64, Y: 32}, 32, true)

	tNow := time.Now()
	result.Print(fmt.Sprintf("%02v:%02v:%02v", tNow.Hour(), tNow.Minute(), tNow.Second()), testfont, 0, 1, result.Bounds(), true, true, false, false)

	return result
}

const (
	PAGE_APP      = 0
	PAGE_APPERR   = 1
	PAGE_STATUS   = 2
	PAGE_KERNEL   = 3
	PAGE_NETWORK  = 4
	PAGE_MOUNTS   = 5
	PAGE_BLOCKDEV = 6
)

func execApp(normalOut chan string, errPrintout chan string) error {
	cmd := exec.Command("./program")
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
	if os.Getpid() == 1 {
		fmt.Printf("-- Starting example init program --\n")

		fmt.Printf("mounting normal set of dirs\n")
		errMount := initializing.MountNormal()
		ctrlaltgo.JamIfErr(errMount)
	}

	pgs := textmonitor.Pages{
		Title:      "otsikko menee 12:34",
		Status:     textmonitor.Normal,
		ActivePage: 0,
		Items:      make([]textmonitor.Page, 7),
	}

	pgs.Items[PAGE_APP] = textmonitor.Page{MenuCaption: "app",
		Content:        "application\ntext\ngoes here",
		ScrollPosition: 0}
	pgs.Items[PAGE_APPERR] = textmonitor.Page{MenuCaption: "apperr",
		Content:        "todo stderr from app",
		ScrollPosition: 0}

	pgs.Items[PAGE_STATUS] = textmonitor.Page{MenuCaption: "sta",
		Content:        "status:TODO\nCPU:\nMEM;\n",
		ScrollPosition: 0}

	pgs.Items[PAGE_KERNEL] = textmonitor.Page{MenuCaption: "ker",
		Content:        "kernelRow0\nkernelRow1\nkernelRow2\nkernelRow3\nkernelRow4\n",
		ScrollPosition: 0,
	}
	pgs.Items[PAGE_NETWORK] = textmonitor.Page{MenuCaption: "net",
		Content:        "TODO network",
		ScrollPosition: 0,
	}
	pgs.Items[PAGE_MOUNTS] = textmonitor.Page{MenuCaption: "mnt",
		Content:        "TODO mounts",
		ScrollPosition: 0,
	}

	pgs.Items[PAGE_BLOCKDEV] = textmonitor.Page{MenuCaption: "blkdev",
		Content:        "TODO block devices",
		ScrollPosition: 0,
	}

	/*
		picrender := gomonochromebitmap.BlockGraphics{
			Clear:      false,
			HaveBorder: false,
			//BorderColor: gomonochromebitmap.FGANSI_YELLOW,
			//TextColor:   gomonochromebitmap.FGANSI_BLUE,
		}*/

	/*
		pic := CreateDemoPic()
		fmt.Printf("%s\n", picrender.ToHalfBlockChars(&pic))
		//return
		sPicCode := picrender.ToHalfBlockChars(&pic)
		os.WriteFile("/tmp/picCode.txt", []byte(sPicCode), 0666)

		pgs.Items[0].Content = sPicCode
		printoutCode := pgs.Items[0].Printout(35, 130)
		os.WriteFile("/tmp/printoutCode.txt", []byte(printoutCode), 0666)

		fmt.Printf("%s\n", printoutCode)
	*/
	//return

	//
	var keyboards inputdev.InputDeviceArr
	for {
		devArr, errParse := inputdev.ParseDevicesFile(inputdev.INPUTDEVICESFILE)
		if errParse != nil {
			fmt.Printf("err:%s\n", errParse)
		}
		keyboards = devArr.GetDefaultKeyboards()
		if len(keyboards) == 0 {
			fmt.Printf("No keyboards found!\n")
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	pKeyboardNumber := flag.Int("i", 0, fmt.Sprintf("index of keyboard, got %v", len(keyboards)))
	flag.Parse()

	if *pKeyboardNumber < 0 || len(keyboards) <= *pKeyboardNumber {
		fmt.Printf("invalid keyboard index, got only %v keyboards\n", len(keyboards))
		return
	}

	kbFileName := keyboards[*pKeyboardNumber].GetDevName()
	fKb, errFKb := os.Open(kbFileName)
	if errFKb != nil {
		fmt.Printf("error reading device %s  err:%s\n", kbFileName, errFKb)
		return
	}

	inputEvents := make(chan inputdev.RawInputEvent, 100)
	go inputdev.ReadEventsToChan(fKb, inputEvents)

	fd := int(os.Stdin.Fd())

	fmt.Printf("making raw\n")
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}

	keyboard := inputdev.InitKeyboard()

	kernelMsgCh := make(chan status.KMsg, 100)
	/*go func() {
		for {
			m := <-kernelMsgCh
			//fmt.Printf("%s\n", m.String())
		}
	}()*/

	monKernel, errOpenKernel := status.OpenKernelMonitor(3000)
	if errOpenKernel != nil {
		fmt.Printf("open err %s\n", errOpenKernel)
		return
	}

	go func() {
		for {
			errRead := monKernel.Read(kernelMsgCh)
			if errRead != nil {
				fmt.Printf("KERNEL READ ERR %s\n", errRead)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()

	var appOutput chan string
	var appErrorOutput chan string
	appOutput = make(chan string, 10000)
	appErrorOutput = make(chan string, 10000)

	go func() {
		for {

			errExec := execApp(appOutput, appErrorOutput)
			if errExec != nil {
				appErrorOutput <- fmt.Sprintf("%s", errExec)
			} else {
				appErrorOutput <- "app exitted with nil error"
			}

			time.Sleep(time.Second)
		}
	}()

	for {
		termCols, termRows, errSize := term.GetSize(fd)
		if errSize != nil {
			fmt.Printf("error size %s\n", errSize)
			return
		}
		fmt.Printf("%s", pgs.Printout(termRows, termCols))

		for 0 < len(appOutput) {
			pgs.Items[PAGE_APP].Content += <-appOutput + "\n"
		}

		for 0 < len(appErrorOutput) {
			pgs.Items[PAGE_APPERR].Content += <-appErrorOutput + "\n"
		}

		//Update data
		switch pgs.ActivePage {
		case PAGE_APP:
			//pic := CreateDemoPic()
			//pgs.Items[PAGE_APP].Content = picrender.ToHalfBlockChars(&pic)
		case PAGE_STATUS:
			sta, errSta := status.GetProcStat()
			if errSta != nil {
				pgs.Items[PAGE_STATUS].Content = fmt.Sprintf("ERROR:%s", errSta)
				break
			}
			pgs.Items[PAGE_STATUS].Content = fmt.Sprintf("CPU: %.2f\n", sta.CPU.CpuPercent())
			for i, cpu := range sta.CPUs {
				pgs.Items[PAGE_STATUS].Content += fmt.Sprintf("CPU%v:%.2f ", i, cpu.CpuPercent())
			}
			pgs.Items[PAGE_STATUS].Content += fmt.Sprintf("\nProcesses:%v  ProcsRunning:%v  ProcsBlocked:%v\n", sta.Processes, sta.ProcsRunning, sta.ProcsBlocked)

		case PAGE_NETWORK:
			ifnames, errListIf := networking.ListInterfaceNames()
			if errListIf != nil {
				pgs.Items[PAGE_NETWORK].Content = fmt.Sprintf("error listing interfaces %s", errListIf)
				break
			}

			pgs.Items[PAGE_NETWORK].Content = ""
			for _, name := range ifnames {
				pgs.Items[PAGE_NETWORK].Content += networking.PrintoutNetInterface(name) + "\n"
			}
		case PAGE_MOUNTS:

			pgs.Items[PAGE_MOUNTS].Content = ""

			mntInfo, errMntInfo := initializing.GetMountInfo()
			if errMntInfo != nil {
				pgs.Items[PAGE_MOUNTS].Content = fmt.Sprintf("error listing mounts %s", errMntInfo)
				break
			}

			//pgs.Items[PAGE_MOUNTS].Content = fmt.Sprintf("\n\n\n*Mounts*\n%s\n*Devices*\n%s\n", mntInfo, blockDevices)
			pgs.Items[PAGE_MOUNTS].Content = "\n\n\n*Mounts*\n"
			for i, m := range mntInfo {
				pgs.Items[PAGE_MOUNTS].Content += fmt.Sprintf("%d:%s\n", i, m)
			}

		case PAGE_BLOCKDEV:

			blockDevices, errBlockDevices := initializing.GetBlockDevices()
			if errBlockDevices != nil {
				pgs.Items[PAGE_BLOCKDEV].Content = fmt.Sprintf("error listing block devices %s", errBlockDevices)
				break
			}
			pgs.Items[PAGE_BLOCKDEV].Content = "*Devices*\n"
			for i, d := range blockDevices {
				pgs.Items[PAGE_BLOCKDEV].Content += fmt.Sprintf("%d:\n%s\n\n", i, d)
			}
		}

		select {
		case event := <-inputEvents:
			keyboard.Update(&event)
			keysDown := keyboard.PressedKeys()
			for _, key := range keysDown {
				switch key {
				case inputdev.KEY_0:
					pgs.ActivePage = 0
				case inputdev.KEY_1:
					pgs.ActivePage = 1
				case inputdev.KEY_2:
					pgs.ActivePage = 2
				case inputdev.KEY_3:
					pgs.ActivePage = 3
				case inputdev.KEY_4:
					pgs.ActivePage = 4
				case inputdev.KEY_5:
					pgs.ActivePage = 5
				case inputdev.KEY_6:
					pgs.ActivePage = 6
				case inputdev.KEY_UP:
					pgs.Items[pgs.ActivePage].ScrollPosition--
				case inputdev.KEY_DOWN:
					pgs.Items[pgs.ActivePage].ScrollPosition++
				case inputdev.KEY_PAGEUP:
					pgs.Items[pgs.ActivePage].ScrollPosition -= termRows
				case inputdev.KEY_PAGEDOWN:
					pgs.Items[pgs.ActivePage].ScrollPosition += termRows
				case inputdev.KEY_SPACE:
					pgs.Items[pgs.ActivePage].ScrollToEnd(termRows)
				}
			}
		case <-kernelMsgCh:
			itms := pgs.Items[PAGE_KERNEL]
			for 0 < len(kernelMsgCh) {
				msg := <-kernelMsgCh
				itms.Content += msg.String() + "\n"
			}
			/*
				pgs.Items[2].Content += msg.String() + "\n"*/
			itms.Content = ""

			for _, itm := range monKernel.History {
				itms.Content += itm.String() + "\n"
			}
			os.WriteFile("content.conn", []byte(itms.Content), 0666)
			itms.ScrollToEnd(termRows)
			pgs.Items[PAGE_KERNEL] = itms

		}

		/*for 0 < len(kernelMsgCh) {
			msg := <-kernelMsgCh
			pgs.Items[2].Content += msg.String() + "\n"
		}*/

		escPress, _ := keyboard.State[inputdev.KEY_ESC]
		if 0 < escPress {
			break
		}

	}

	term.Restore(fd, oldState)
	//time.Sleep(time.Second * 10)

	fmt.Printf("kernel scroll process %v\n", pgs.Items[2].ScrollPosition)
	fmt.Printf("hist %v/%v\n", len(monKernel.History), cap(monKernel.History))
	fmt.Printf("rows=%v\n", len(strings.Split(pgs.Items[2].Content, "\n")))

	//fmt.Printf("termina: %v x %v\n", termRows, termCols)
	//fmt.Printf("printout row count %v\n", len(strings.Split(s, "\n")))
}
