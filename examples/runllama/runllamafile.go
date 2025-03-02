/*
Example how to run program written in C/C++ with dynamic libraries
llama.cpp is just for example
*/

package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
	_ "time/tzdata" //Smart thing to have up to date tzdata here

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/hjkoskel/ctrlaltgo/networking"
)

const INTERFACENAME = "eth0"
const HOSTNAMEOFSYSTEM = "simplesrv"
const EXTRACTDIRECTORY_LLAMA = "/software/llamacpp" //Both binary and dependencies
const LLAMABINARYNAME = "llama-server"

//Doing better than AppImage

//go:embed content.tar.gz
var contentTarGz []byte

/*
try something small
https://huggingface.co/unsloth/SmolLM2-135M-Instruct-GGUF/tree/main
*/

//go:embed SmolLM2-135M-Instruct-Q4_K_M.gguf
var modelbytes []byte

const MODEFILEONRUNTIME = "/tmp/model.gguf"

func ExtractTo(targetDir string, targzcontent []byte) ([]string, error) {
	filenamelist := []string{}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return filenamelist, fmt.Errorf("failed to create target directory: %v\n", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(targzcontent))
	if err != nil {
		return filenamelist, fmt.Errorf("failed to create gzip reader: %v\n", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return filenamelist, fmt.Errorf("failed to read tar header: %v", err)
		}

		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			fmt.Printf("created dir %s\n", targetPath)
			if err := os.MkdirAll(targetPath, 0777); err != nil {
				return filenamelist, fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			fmt.Printf("writing file %s\n", targetPath)
			filenamelist = append(filenamelist, targetPath)
			// Create file
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return filenamelist, fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				return filenamelist, fmt.Errorf("failed to write file content: %v", err)
			}
			errClose := file.Close()
			if errClose != nil {
				return filenamelist, fmt.Errorf("error closing file %s err:%s", targetPath, errClose)
			}
			fmt.Printf("set all permissions\n")
			errChmod := os.Chmod(targetPath, 0777)
			if errChmod != nil {
				return filenamelist, fmt.Errorf("error setting chmod on file %s err:%s", targetPath, errChmod)
			}
		default:
			return filenamelist, fmt.Errorf("unsupported file type: %v", header.Typeflag)

		}
	}
	return filenamelist, nil
}

func pickFile(fnames []string, pathPrefix string, fileNamePrefix string) string {
	for _, fname := range fnames {
		if strings.HasPrefix(path.Base(fname), fileNamePrefix) {
			if pathPrefix == "" || strings.HasPrefix(fname, pathPrefix) {
				return fname
			}
		}
	}
	return ""
}

// RunCommand executes a command and streams stdout and stderr in real time
func RunCommand(dynlinker string, executableName string, libPaths []string, par ...string) error {
	parArr := []string{}
	for _, s := range libPaths {
		parArr = append(parArr, "--library-path")
		parArr = append(parArr, s)
	}

	parArr = append(parArr, executableName)
	parArr = append(parArr, par...)
	fmt.Printf("par array is %s\n", strings.Join(parArr, " "))

	cmd := exec.Command(dynlinker, parArr...)
	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	// Stream output concurrently
	go streamOutput(stdout, "STDOUT")
	go streamOutput(stderr, "STDERR")

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// streamOutput reads from an io.Reader and prints it line by line
func streamOutput(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("[%s] %s\n", prefix, scanner.Text())
	}
}

func main() {
	if os.Getpid() == 1 {
		errMount := initializing.MountNormal()
		ctrlaltgo.JamIfErr(errMount)
		errSetHostname := initializing.SetHostname(HOSTNAMEOFSYSTEM) //Important when having network
		ctrlaltgo.JamIfErr(errSetHostname)

		hostnameByKernel, errHostname := os.Hostname()
		ctrlaltgo.JamIfErr(errHostname)

		fmt.Printf("hostname by kernel:%s\n", hostnameByKernel)

		//Doing network setup
		fmt.Printf("Bring %v up\n", INTERFACENAME)

		errWaitInterf := networking.WaitInterface(INTERFACENAME, time.Second*30, time.Second) //Raspberry pi delay?
		ctrlaltgo.JamIfErr(errWaitInterf)

		errUp := networking.SetLinkUp(INTERFACENAME, true)
		ctrlaltgo.JamIfErr(errUp)
		fmt.Printf("Checking carrier on %s ....", INTERFACENAME)
		for {

			haveCarr, errCarr := networking.Carrier(INTERFACENAME)
			ctrlaltgo.JamIfErr(errCarr)
			if haveCarr {
				break
			}
		}
		fmt.Printf("... have carrier\n")

		//Run in goroutine?
		ipSettings, errDhcp := networking.GetDHCP(HOSTNAMEOFSYSTEM, INTERFACENAME)
		ctrlaltgo.JamIfErr(errDhcp)

		fmt.Printf("GOT IP settings %#v\n", ipSettings)

		errApplyIp := ipSettings.ApplyToInterface(INTERFACENAME, 1)
		ctrlaltgo.JamIfErr(errApplyIp)
	}

	fmt.Printf("**** STARTING GOLANG PROG HERE ******\n")

	extractedFiles, errExtract := ExtractTo(EXTRACTDIRECTORY_LLAMA, contentTarGz)

	//TWO most important
	llamaServerFilename := pickFile(extractedFiles, "", LLAMABINARYNAME)
	if llamaServerFilename == "" {
		ctrlaltgo.JamIfErr(fmt.Errorf("llama server binary not found"))
	}
	dynLinkerFilename := pickFile(extractedFiles, "", "ld-linux-")
	if dynLinkerFilename == "" {
		ctrlaltgo.JamIfErr(fmt.Errorf("dynamic linker ld-linux-* binary not found"))
	}

	ctrlaltgo.JamIfErr(errExtract)
	fmt.Printf("Extracted...going to run llama\n")
	fmt.Printf("dynlinker=%s\nserver=%s\n", dynLinkerFilename, llamaServerFilename)

	errWriteModel := os.WriteFile(MODEFILEONRUNTIME, modelbytes, 0666)
	if errWriteModel != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error writing model data from binary"))
	}

	errRun := RunCommand(dynLinkerFilename, llamaServerFilename, []string{EXTRACTDIRECTORY_LLAMA}, "--port", "8888", "--host", "0.0.0.0", "-m", MODEFILEONRUNTIME)
	fmt.Printf("run command is over")
	ctrlaltgo.JamIfErr(errRun)

	fmt.Printf("errrun is nil\n")
}
