/*
 */
package main

import (
	"bytes"
	"debug/elf"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/hjkoskel/ctrlaltgo/initializing"
)

const (
	SERVERTCPPORT = 4242
)

//go:embed webmaint
var staticWebMainananceGui embed.FS

func safeWrite(targetFileName string, content []byte) error {
	tmpFilename := targetFileName + "_tmp"

	f, cErr := os.Create(tmpFilename)
	if cErr != nil {
		return fmt.Errorf("error creating tmp file %s err:%s", tmpFilename, cErr)
	}

	n, wErr := f.Write(content)
	if wErr != nil {
		return fmt.Errorf("error writing content err:%s", wErr)
	}
	if n != len(content) {
		return fmt.Errorf("tried write %v bytes, wrote %v", len(content), n)
	}

	errSync := f.Sync()
	if errSync != nil {
		return fmt.Errorf("error syncing tmp file err:%s", errSync)
	}

	errClose := f.Close()
	if errClose != nil {
		return fmt.Errorf("error closing tmp file %s err:%s", tmpFilename, errClose)
	}

	errRen := os.Rename(tmpFilename, targetFileName)
	if errRen != nil {
		return fmt.Errorf("error renaming %s to %s err:%s", tmpFilename, targetFileName, errRen)
	}
	return nil
}

// Writes only if this is valid
func SafeWriteElfBinary(fname string, binData []byte, machineWanted elf.Machine) error {
	// Parse the ELF file

	elfFile, err := elf.NewFile(bytes.NewReader(binData))
	if err != nil {
		return fmt.Errorf("file is not a valid ELF: %v", err)
	}
	fmt.Printf("binary arch is %#s\n", elfFile.Machine)
	// Check if the machine architecture is ARM64
	if elfFile.Machine != machineWanted {
		return errors.New("file is not an ARM64 ELF executable")
	}
	return safeWrite(fname, binData)
}

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

	errSave := SafeWriteElfBinary(target, binData, elf.EM_AARCH64) //64bits
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

func MaintananceServer(programQueue chan string, stdinCh chan string, stderrch chan string, crashch chan error) error {

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
		for s := range stdinCh { //TODO LIMIT ROW COUNT?
			stdoutRows = append(stderrRows, s) //TODO add timestamp metadata?
			fmt.Printf("ERR! %s\n", s)
		}
	}()

	go func() {
		for e := range crashch {
			crashed = e.Error() //TODO timestampping?
		}
	}()

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

	//TODO separe function when this works
	http.HandleFunc("POST /progTEMP", func(w http.ResponseWriter, r *http.Request) {
		HandleProgramWrite(w, r, TMPPROGRAM, programQueue)

		/*
			fmt.Printf("WANTING TO PROGRAM %s\n", TMPPROGRAM)
			binData, errRead := GetMultipartBytes(r, (10<<20)*256)
			if errRead != nil {
				w.Write([]byte(fmt.Sprintf("error getting message body err:%s", errRead)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			programQueue <- PAUSEPROG
			programQueue <- PAUSEPROG
			programQueue <- PAUSEPROG
			programQueue <- PAUSEPROG //Channel is clearing when it is waiting idle

			errSave := SafeWriteElfBinary(TMPPROGRAM, binData, elf.EM_AARCH64) //64bits
			if errSave != nil {
				w.Write([]byte(fmt.Sprintf("error saving %s err:%s", TMPPROGRAM, errSave)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Printf("Going to start %s\n", TMPPROGRAM)
			programQueue <- TMPPROGRAM
			fmt.Printf("Started....\n")
			fmt.Fprint(w, "ack")
		*/
	})

	http.HandleFunc("POST /prog", func(w http.ResponseWriter, r *http.Request) {
		HandleProgramWrite(w, r, PROGRAM, programQueue)
		/*binData, errRead := GetMultipartBytes(r, (10<<20)*256)
		if errRead != nil {
			w.Write([]byte(fmt.Sprintf("error getting message body err:%s", errRead)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG
		programQueue <- PAUSEPROG //Channel is clearing when it is waiting idle

		errSave := SafeWriteElfBinary(PROGRAM, binData, elf.EM_AARCH64) //64bits
		if errSave != nil {
			w.Write([]byte(fmt.Sprintf("error saving %s err:%s", PROGRAM, errSave)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("Going to start %s\n", PROGRAM)
		programQueue <- PROGRAM
		fmt.Printf("Started....\n")
		fmt.Fprint(w, "ack")*/
	})

	fmt.Printf("starting server on %v\n", SERVERTCPPORT)

	return http.ListenAndServe(fmt.Sprintf(":%v", SERVERTCPPORT), nil)
}
