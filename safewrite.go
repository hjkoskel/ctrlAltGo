package ctrlaltgo

import (
	"bytes"
	"debug/elf"
	"fmt"
	"os"
	"path"
)

func SafeWrite(targetFileName string, content []byte) error {
	tmpFilename := targetFileName + "_tmp"
	os.Mkdir(path.Dir(targetFileName), 0777)

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
	// Check if the machine architecture is ARM64
	if elfFile.Machine != machineWanted {
		return fmt.Errorf("file is not an %w executable", machineWanted)
	}
	return SafeWrite(fname, binData)
}
