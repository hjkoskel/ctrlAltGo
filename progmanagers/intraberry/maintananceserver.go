/*
 */
package main

import (
	"debug/elf"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/hjkoskel/ctrlaltgo/status"
)

const (
	SERVERTCPPORT = 4242
)

//go:embed webmaint
var staticWebMainananceGui embed.FS

func GetMultipartBytes(r *http.Request, maxMemory int) ([]byte, error) {
	err := r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		return nil, fmt.Errorf("unable to parse form err:%s", err)
	}
	fmt.Printf("parsed multipart")
	// Retrieve the file from the form data
	file, _, errFormFile := r.FormFile("program")
	if errFormFile != nil {
		return nil, fmt.Errorf("unable to retrieve file, err:%s", errFormFile)
	}
	defer file.Close()
	fmt.Printf("got file. reading all\n")

	return io.ReadAll(file)
}

var stdoutRows []string
var stderrRows []string
var crashed string
var machineArchNow elf.Machine

func HandleProgramWrite(w http.ResponseWriter, r *http.Request, target string, queue chan string) {
	fmt.Printf("WANTING TO PROGRAM %s\n", target)
	binData, errRead := GetMultipartBytes(r, (10<<20)*256)
	if errRead != nil {
		w.Write([]byte(fmt.Sprintf("error getting message body err:%s", errRead)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	queue <- PAUSEPROG
	queue <- PAUSEPROG
	queue <- PAUSEPROG
	queue <- PAUSEPROG //Channel is clearing when it is waiting idle

	errSave := ctrlaltgo.SafeWriteElfBinary(target, binData, machineArchNow) //64bits
	if errSave != nil {
		w.Write([]byte(fmt.Sprintf("error saving %s err:%s", target, errSave)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	os.Chmod(target, 0777)
	fmt.Printf("Going to start %s to queue\n", target)
	queue <- target
	fmt.Fprint(w, "ack")
}

func MaintananceServer(programQueue chan string, stdinCh chan string, stderrch chan string, crashch chan error, kernelMsgCh chan status.KMsg) error {
	var errArch error
	machineArchNow, errArch = ctrlaltgo.GetCurrentMachine()
	if errArch != nil {
		return errArch
	}

	statifscontent, errstaticconten := fs.Sub(staticWebMainananceGui, "webmaint")
	if errstaticconten != nil {
		return errstaticconten
	}

	fs := http.FileServer(http.FS(statifscontent))
	http.Handle("/", fs)

	go func() {
		for s := range stdinCh { //TODO LIMIT ROW COUNT?
			stdoutRows = append(stdoutRows, s) //TODO add timestamp metadata?
			fmt.Printf("%s\n", s)
		}
	}()

	go func() {
		for s := range stderrch { //TODO LIMIT ROW COUNT?
			stderrRows = append(stderrRows, s) //TODO add timestamp metadata?
		}
	}()

	go func() {
		for e := range crashch {
			crashed = e.Error() //TODO timestampping?
		}
	}()

	var kernelMessageBuffer string
	//TODO better formatting?
	go func() {
		for msg := range kernelMsgCh {
			kernelMessageBuffer += msg.String() + "\n"
		}
	}()
	http.HandleFunc("GET /kmsg", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, kernelMessageBuffer)
	})

	//https://go.dev/blog/routing-enhancements

	http.HandleFunc("GET /stdout", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Join(stdoutRows, "\n"))
	})

	http.HandleFunc("GET /stderr", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Join(stderrRows, "\n"))
	})

	http.HandleFunc("GET /chrashmsg", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, crashed)
	})

	http.HandleFunc("GET /restart", func(w http.ResponseWriter, r *http.Request) {
		programQueue <- ""
		fmt.Fprint(w, "ack")
	})

	http.HandleFunc("GET /reboot", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("going to reboot, pausing program\n")
		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG
		fmt.Printf("NOW rebooting!\n")
		errReboot := initializing.Reboot()
		fmt.Printf("Error reboot %s\n", errReboot)
		if errReboot != nil {
			fmt.Fprint(w, errReboot.Error())
		}
	})

	http.HandleFunc("GET /cpu", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "todo cpu")
	})

	http.HandleFunc("GET /mounts", func(w http.ResponseWriter, r *http.Request) {
		mntInfo, errMntInfo := initializing.GetMountInfo()
		if errMntInfo != nil {
			w.Write([]byte(fmt.Sprintf("error listing mounts %s", errMntInfo)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		byt, _ := json.MarshalIndent(mntInfo, "", " ")

		w.Write(byt)
	})

	http.HandleFunc("GET /blockdevices", func(w http.ResponseWriter, r *http.Request) {
		blockDevices, errBlockDevices := initializing.GetBlockDevices()
		if errBlockDevices != nil {
			w.Write([]byte(fmt.Sprintf("error listing blockDevices %s", errBlockDevices)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		byt, _ := json.MarshalIndent(blockDevices, "", " ")
		w.Write(byt)
	})

	http.HandleFunc("GET /procinfos", func(w http.ResponseWriter, r *http.Request) {
		info, errInfo := status.ReadProcessInfos("/proc")
		if errInfo != nil {
			w.Write([]byte(fmt.Sprintf("error proc infos %s", errInfo)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//byt, _ := json.MarshalIndent(info, "", " ")
		//w.Write(byt)
		w.Write([]byte(info.String())) //TODO json later?
	})

	http.HandleFunc("GET /filehandleusage", func(w http.ResponseWriter, r *http.Request) {
		usage, errUsage := status.GetFileHandlesUsage()
		if errUsage != nil {
			w.Write([]byte(fmt.Sprintf("error proc infos %s", errUsage)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(usage.String()))
	})

	http.HandleFunc("GET /procstat", func(w http.ResponseWriter, r *http.Request) {
		sta, errSta := status.GetProcStat()
		if errSta != nil {
			w.Write([]byte(fmt.Sprintf("error listing proc stat %s", errSta)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		byt, _ := json.MarshalIndent(sta, "", " ")
		w.Write(byt)
	})

	//TODO separe function when this works
	http.HandleFunc("POST /progTEMP", func(w http.ResponseWriter, r *http.Request) {
		HandleProgramWrite(w, r, TMPPROGRAM, programQueue)
	})

	http.HandleFunc("POST /prog", func(w http.ResponseWriter, r *http.Request) {
		HandleProgramWrite(w, r, PROGRAM, programQueue)
	})

	//Very brutal way, but when developing this is ok to do. Never so straightforward on production!
	http.HandleFunc("POST /initramfsupdate", func(w http.ResponseWriter, r *http.Request) {
		//extractCPIOArchive(
		binData, errRead := GetMultipartBytes(r, (10<<20)*256)
		if errRead != nil {
			w.Write([]byte(fmt.Sprintf("error getting message body err:%s", errRead)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/*
			TODO UPDATE
			valid, errValidate := validateInitramfs(bytes.NewReader(binData), machineArchNow)
			if errValidate != nil {
				w.Write([]byte(fmt.Sprintf("err validate %s", errValidate)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if !valid {
				w.Write([]byte("package is not valid initramfs"))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Printf("IS VALID, going to write initramfs\n")

		*/

		targetFileName := path.Join(MNT_BOOTCARD, "initramfs")
		fmt.Printf("GOT %v bytes, gointo write %s\n", len(binData), targetFileName)

		errSave := ctrlaltgo.SafeWrite(targetFileName, binData)
		if errSave != nil {
			w.Write([]byte(fmt.Sprintf("error saving err:%s", errSave)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	})

	fmt.Printf("starting server on %v\n", SERVERTCPPORT)

	return http.ListenAndServe(fmt.Sprintf(":%v", SERVERTCPPORT), nil)
}
