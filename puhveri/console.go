package puhveri

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const (
	KDSETMODE   = 0x4b3a
	KD_GRAPHICS = 0x1
)

// a very hacky way
func ConsoleToGraphics() error {
	f, err := os.OpenFile("/dev/console", os.O_RDWR, 0)
	if err != nil {
		return err
	}

	if err := unix.IoctlSetInt(int(f.Fd()), KDSETMODE, KD_GRAPHICS); err != nil {
		return fmt.Errorf("KDSETMODE: %v", err)
	}
	return nil
}
