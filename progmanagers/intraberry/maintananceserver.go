/*
 */
package main

import (
	"debug/elf"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/hjkoskel/ctrlaltgo/inputdev"
	"github.com/hjkoskel/ctrlaltgo/status"
	"golang.org/x/sys/unix"
)

const (
	SERVERTCPPORT = 4242
)

//go:embed i2cdevices_min.json
var i2cDeviceDatabaseBytes []byte //I2CDeviceDatabase

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

type FileBrowserViewData struct {
	UpdatedTimeAndDate string
	De                 status.DirectoryEntry
	Browse             status.TreeOpening //Directory structure on side menu
	PreviewText        string             //Allows to preview content on extra column
	SelectedFileName   string
	IsOneLineFile      bool
}

func (d FileBrowserViewData) SideTree() template.HTML {
	var sb strings.Builder
	sb.WriteString("<ul>\n")
	sb.WriteString(d.Browse.ToUlList())
	sb.WriteString("\n</ul>")
	return template.HTML(sb.String())
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

	//I2c data for translating device names
	var i2cdevicedb I2CDeviceDatabase
	errInternalI2CdbParseErr := json.Unmarshal(i2cDeviceDatabaseBytes, &i2cdevicedb)
	if errInternalI2CdbParseErr != nil {
		return fmt.Errorf("internal i2c database parse error %s", errInternalI2CdbParseErr)
	}

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

	http.HandleFunc("GET /cpuinfo", func(w http.ResponseWriter, r *http.Request) {
		byt, errRead := os.ReadFile("/proc/cpuinfo")
		if errRead != nil {
			w.Write([]byte(errRead.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(byt)
	})

	http.HandleFunc("GET /cpu", func(w http.ResponseWriter, r *http.Request) {
		info, errCpu := initializing.GetCpuinfo("/proc/cpuinfo")
		if errCpu != nil {
			w.Write([]byte(fmt.Sprintf("errCPU %s\n", errCpu)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		info.Commonize()
		byt, _ := json.MarshalIndent(info, "", " ")
		w.Write(byt)
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
		w.Write([]byte(blockDevices.String()))
	})

	//procPartitions for testing... major and minor numbers are needed?
	http.HandleFunc("GET /procpartitions", func(w http.ResponseWriter, r *http.Request) {
		data, errParse := initializing.ParseProcPartitions()
		if errParse != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error parsin /proc/partitions %s", errParse)))
			return
		}
		byt, _ := json.MarshalIndent(data, "", " ")
		w.Write(byt)
	})

	http.HandleFunc("GET /procdevices", func(w http.ResponseWriter, r *http.Request) { //For debug/getting major numbers of devices
		procdev, errProcdev := status.LoadProcDevices()
		if errProcdev != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error proc devices %s", errProcdev)))
			return
		}
		byt, _ := json.MarshalIndent(procdev, "", " ")
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

	//Get environment
	http.HandleFunc("GET /env", func(w http.ResponseWriter, r *http.Request) {
		env, errEnv := initializing.GetEnvs()
		if errEnv != nil {
			w.Write([]byte(fmt.Sprintf("%s", errEnv)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(env.String()))
	})

	//Text format?
	http.HandleFunc("POST /env", func(w http.ResponseWriter, r *http.Request) {
		byt, errBody := io.ReadAll(r.Body)
		if errBody != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("%s", errBody)))
			return
		}

		keystrings, errParse := initializing.ParseEnvs(string(byt))
		if errParse != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("parse err %s, (400)", errParse)))
			return
		}

		os.Clearenv()
		errSet := keystrings.SetEnvs()
		if errSet != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error setting environment variables %s", errSet)))
			return
		}
		env, errEnv := initializing.GetEnvs()
		if errEnv != nil {
			w.Write([]byte(fmt.Sprintf("%s", errEnv)))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte(env.String()))
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

	http.HandleFunc("GET /meminfo", func(w http.ResponseWriter, r *http.Request) {
		sta, errSta := status.ReadMemInfo(status.MEMINFOFILE)
		if errSta != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error getting meminfo %s", errSta)))
			return
		}
		byt, _ := json.MarshalIndent(sta, "", " ")
		w.Write(byt)
	})

	http.HandleFunc("GET /modules", func(w http.ResponseWriter, r *http.Request) {
		modu, errModu := status.ReadKernelModuleStates("/proc/modules")
		if errModu != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error getting meminfo %s", errModu)))
			return
		}
		byt, errIdent := json.MarshalIndent(modu, "", " ")
		if errIdent != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("err marshaling %s", errIdent)))
			return
		}
		w.Write(byt)
	})

	http.HandleFunc("GET /modulesavailable", func(w http.ResponseWriter, r *http.Request) { //List available modules
		dircontent, errDir := os.ReadDir(KERNELDRIVERDIR)
		if errDir != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error reading dir %s", errDir)))
			//w.Write([]byte("[\"none\"]"))
			return
		}
		names := []string{}
		for _, f := range dircontent {
			if f.IsDir() {
				continue
			}
			if strings.HasSuffix(f.Name(), ".ko") { //TODO expand suffixes if when compression is supported
				names = append(names, f.Name())
			}
		}
		byt, _ := json.MarshalIndent(names, "", " ")
		w.Write(byt)
	})

	http.HandleFunc("GET /loadmodule/{modulename}", func(w http.ResponseWriter, r *http.Request) {
		modulePath := path.Join(KERNELDRIVERDIR, r.PathValue("modulename"))
		fmt.Printf("going to load %s\n", modulePath)
		byt, errRead := os.ReadFile(modulePath)
		if errRead != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("file %s load fail err:%s", modulePath, errRead)))
			return
		}
		errLoadKernelModule := unix.InitModule(byt, "")

		//errLoadKernelModule := initializing.LoadKernelModule(modulePath, "")
		if errLoadKernelModule != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("loading %s failed:%s", r.PathValue("modulename"), errLoadKernelModule)))
			return
		}
		w.Write([]byte(fmt.Sprintf("LOADED %s", modulePath)))
	})

	http.HandleFunc("GET /unloadmodule/{modulename}", func(w http.ResponseWriter, r *http.Request) {
		err := unix.DeleteModule(r.PathValue("modulename"), unix.O_NONBLOCK)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("unloading %s failed:%s", r.PathValue("modulename"), err)))
			return
		}
		w.Write([]byte(fmt.Sprintf("UNLOADED %s", r.PathValue("modulename"))))
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

	http.HandleFunc("GET /browse/{fname...}", func(w http.ResponseWriter, r *http.Request) {
		params, _ := url.ParseQuery(r.URL.RawQuery)
		previewFile := params.Get("preview")

		fname := r.PathValue("fname")
		fileInfo, err := os.Stat(fname)

		if err == nil {
			if !fileInfo.IsDir() {
				http.ServeFile(w, r, fname)
				return
			}
		}
		rootfilename := "/" + fname
		dirEntry, _ := status.ReadDirectoryEntry(rootfilename, false)

		//errTemplateExecute := basicDirHTMLTemplate.Execute(w,  dirEntry)
		tNow := time.Now()

		outData := FileBrowserViewData{
			UpdatedTimeAndDate: fmt.Sprintf("%02d:%02d:%02d: %d.%d.%d", tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Day(), tNow.Month(), tNow.Year()),
			De:                 dirEntry,
			Browse:             status.ReadOpeningFromDir(fname),
		}
		if len(previewFile) != 0 {
			outData.SelectedFileName = path.Join(rootfilename, previewFile)
			byt, _ := os.ReadFile(outData.SelectedFileName)
			if 0 < len(byt) {
				outData.PreviewText = string(byt)
			}
			outData.IsOneLineFile = len(strings.Split(strings.TrimSpace(outData.PreviewText), "\n")) < 2
		}

		errTemplateExecute := basicDirHTMLTemplate.Execute(w, outData)
		if errTemplateExecute != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error executing template %s", errTemplateExecute)))
			return
		}
	})

	http.HandleFunc("GET /opentree/{fname...}", func(w http.ResponseWriter, r *http.Request) {
		result := status.ReadOpeningFromDir(r.PathValue("fname"))
		byt, _ := json.MarshalIndent(result, "", " ")
		w.Write(byt)
	})

	http.HandleFunc("GET /i2cscan/", func(w http.ResponseWriter, r *http.Request) {
		buslist, err := ListI2CBusses()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		resultbytes, _ := json.MarshalIndent(buslist, "", " ")
		w.Write(resultbytes)
	})

	http.HandleFunc("GET /i2cscan/{devicefile}", func(w http.ResponseWriter, r *http.Request) { //List available modules
		devfile := path.Join("/dev/", r.PathValue("devicefile"))
		lst, lstInUse, errlst := ScanI2C(devfile)
		if errlst != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("error reading %s err:%s", devfile, errlst)))
			return
		}
		report := i2cdevicedb.CreateReport(lst, lstInUse)
		byt, errMarsh := json.MarshalIndent(report, "", " ")
		if errMarsh != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("marshal error %s", errMarsh)))
			return
		}
		w.Write(byt)
	})

	http.HandleFunc("GET /inputdevices", func(w http.ResponseWriter, r *http.Request) {
		devArr, errParse := inputdev.ParseDevicesFile(inputdev.INPUTDEVICESFILE)
		if errParse != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("err parse devices file err:%s", &errParse)))
		}
		byt, _ := json.Marshal(devArr)
		w.Write(byt)
	})

	http.HandleFunc("POST /writeSingleLineFile", func(w http.ResponseWriter, r *http.Request) {
		errParseForm := r.ParseForm()
		if errParseForm != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("error parsing form %s", errParseForm)))
			return
		}

		// Retrieve form values
		filename := r.FormValue("filename")
		contentline := r.FormValue("contentline")
		fmt.Printf("write to %s line:%s\n", filename, contentline)
		errWrite := os.WriteFile(filename, []byte(contentline), 0666)
		if errWrite != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("writing %s to %s failed err:%s", contentline, filename, errWrite)))
			return
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

	})

	errInit := initDirGenerator()
	if errInit != nil {
		return fmt.Errorf("dir template err init %w", errInit)

	}

	fmt.Printf("starting server on %v\n", SERVERTCPPORT)

	return http.ListenAndServe(fmt.Sprintf(":%v", SERVERTCPPORT), nil)
}
