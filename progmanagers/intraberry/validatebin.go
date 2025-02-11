/*
function validating binary files
*/
package main

import (
	"debug/elf"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/hjkoskel/ctrlaltgo/deployment"
)

/*
// isARM64ELF checks if a given byte slice is an ARM64 ELF executable.
func isARM64ELF(data []byte) (bool, error) {
	elfFile, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("not a valid ELF file: %v", err)
	}
	defer elfFile.Close()

	if elfFile.Machine != elf.EM_AARCH64 {
		return false, errors.New("not an ARM64 ELF executable")
	}

	return true, nil
}*/

func isGzipCompressedFile(file *os.File) (bool, error) {
	// Read the first two bytes to check for the gzip magic number
	magic := make([]byte, 2)
	_, err := file.Read(magic)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}

	// Reset the file pointer to the beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return false, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	// Check for gzip magic number (0x1F 0x8B)
	return magic[0] == 0x1F && magic[1] == 0x8B, nil
}

/*
gzReader, err := gzip.NewReader(file)
*/

// Check if the file is gzip compressed
/*if isGzPacked {
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return false, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()
	reader = gzReader
}*/

//extractCPIOArchive(bytes.NewReader(binData), outputDir string)

func validateInitramfs(reader io.Reader, wantedMachine elf.Machine) (bool, error) {
	dirname, errMkdir := os.MkdirTemp("/tmp", "validate")
	if errMkdir != nil {
		return false, errMkdir
	}
	defer os.RemoveAll(dirname)

	fmt.Printf("GOING TO TEST EXTRACT TO %s\n", dirname)

	errExtract := deployment.ExtractCPIOArchive(reader, dirname)
	if errExtract != nil {
		return false, fmt.Errorf("can not extract %w", errExtract)
	}
	fmt.Printf("EXTrACTION OK\n")
	//time.Sleep(time.Hour)

	entries, errDir := os.ReadDir(dirname)
	if errDir != nil {
		return false, fmt.Errorf("error dir %w", errDir)
	}
	namelist := []string{}
	for _, entry := range entries {
		/*
			fmt.Printf("name is %s\n", entry.Name())
			time.Sleep(time.Second * 20)
		*/
		if entry.Name() == "init" {
			f, errOpen := os.Open(path.Join(dirname, entry.Name()))
			if errOpen != nil {
				return false, fmt.Errorf("can not open initramfs %w", errOpen)
			}

			elfFile, err := elf.NewFile(f)
			if err != nil {
				return false, nil //Not internal proglem
			}
			f.Close()
			return elfFile.Machine == wantedMachine, nil

		}
		namelist = append(namelist, entry.Name())
	}

	return false, fmt.Errorf("initramfs not found in dir [%s]", strings.Join(namelist, ","))

}

// validateInitramfs checks if the initramfs file is valid for ARM64. Passes correct reader depending on is it gz zipped or not

/*
func validateInitramfs(reader io.Reader, wantedMachine elf.Machine) (bool, error) {

	// Parse the cpio archive
	cpioReader := cpio.NewReader(reader)
	for {
		header, err := cpioReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return false, fmt.Errorf("failed to read cpio archive: %v", err)
		}

		// Check if the file is a regular file
		if header.Mode&cpio.TypeReg == 0 {
			continue // Skip non-regular files
		}

		// Read the file content
		fileData, err := io.ReadAll(cpioReader)
		if err != nil {
			return false, fmt.Errorf("failed to read file from cpio archive: %v", err)
		}

		// Check if the file is an ARM64 ELF executable
		isARM64, err := isARM64ELF(fileData)
		if err == nil && isARM64 {
			fmt.Printf("Found ARM64 ELF executable: %s\n", header.Name)
			return true, nil
		}
	}

	return false, errors.New("no ARM64 ELF executables found in initramfs")
}
*/
