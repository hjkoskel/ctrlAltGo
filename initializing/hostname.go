package initializing

import (
	"os"
	"strings"
	"syscall"
)

func SetHostname(hostname string) error {
	hn := []byte(strings.TrimSpace(hostname))
	errHost := os.WriteFile("/etc/hostname", hn, 0666)
	if errHost != nil {
		return errHost
	}
	return syscall.Sethostname(hn)
}
