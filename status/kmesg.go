/*
kmsg log monitorin
monitor messages or
*/
package status

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Facility int

const (
	Kern Facility = iota
	User
	Mail
	Daemon
	Auth
	Syslog
	Lpr
	News
	Uucp
	Cron
	AuthPriv
	Local0
	Local1
	Local2
	Local3
	Local4
	Local5
	Local6
	Local7
)

func (f Facility) String() string {
	return [...]string{
		"kern", "user", "mail", "daemon",
		"auth", "syslog", "lpr", "news", "uucp",
		"cron", "authpriv",
		"local0", "local1", "local2", "local3",
		"local4", "local5", "local6", "local7",
	}[f]
}

type Level int

const (
	Emerg Level = iota
	Alert
	Crit
	Err
	Warning
	Notice
	Info
	Debug
)

func (p Level) String() string {
	return [...]string{"emerg", "alert", "crit", "err", "warning", "notice", "info", "debug"}[p]
}

type KMsg struct {
	Level    Level    // SYSLOG lvel
	Facility Facility // SYSLOG facility
	Seq      uint64   // Message sequence number
	//TsUsec     int64             // Timestamp in microsecond
	Dur        time.Duration
	Caller     string            // Message caller
	IsFragment bool              // This message is a fragment of an early message which is not a fragment
	Text       string            // Log text
	DeviceInfo map[string]string // Device info
}

func (a KMsg) String() string {
	/*arr := []string{}
	for key, value := range a.DeviceInfo {
		arr = append(arr, fmt.Sprintf(" %s=%s ", key, value))
	}*/
	/*return fmt.Sprintf("%s/%s:%v %s@%s %s [%s]", a.Level, a.Facility, a.Seq,
	a.Caller, a.Dur, a.Text, strings.Join(arr, ","))*/
	return fmt.Sprintf("%s/%s:%v %s@%s %s", a.Level, a.Facility, a.Seq,
		a.Caller, a.Dur, a.Text)

}

//“level,sequnum,timestamp;<message text>\n”

func parseKmsg(s string) (KMsg, error) {
	msg := KMsg{}
	ab := strings.Split(s, ";")
	if len(ab) < 2 {
		return KMsg{}, fmt.Errorf("invalid line missing ;  row:%s", s)
	}
	arr := strings.Split(ab[0], ",")

	for index, prefix := range arr {
		switch index {
		case 0:
			val, _ := strconv.ParseUint(string(prefix), 10, 64)
			msg.Level = Level(val & 7)
			msg.Facility = Facility(val >> 3)
		case 1:
			val, _ := strconv.ParseUint(string(prefix), 10, 64)
			msg.Seq = val
		case 2:
			val, _ := strconv.ParseInt(string(prefix), 10, 64)
			msg.Dur = time.Duration(val) * time.Microsecond
		case 3:
			msg.IsFragment = prefix[0] != '-'
		case 4:
			msg.Caller = string(prefix)
		}
	}
	txtRows := strings.Split(strings.Join(ab[1:], ";"), "\n")
	msg.Text = txtRows[0]
	msg.DeviceInfo = make(map[string]string)
	for i, row := range txtRows {
		if i == 0 {
			continue
		}
		keyvalue := strings.Split(row, "=")
		if len(keyvalue) != 2 {
			msg.DeviceInfo[strings.TrimSpace(keyvalue[0])] = ""
		} else {
			msg.DeviceInfo[strings.TrimSpace(keyvalue[0])] = strings.TrimSpace(keyvalue[1])
		}
	}
	return msg, nil
}

type KernelMonitor struct {
	f       *os.File
	conn    syscall.RawConn
	History []KMsg //TODO capacity limit
	Cap     int
}

func OpenKernelMonitor(nMaxEntries int) (KernelMonitor, error) {
	file, err := os.OpenFile("/dev/kmsg", syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return KernelMonitor{}, fmt.Errorf("error opening /dev/kmsg: %v", err)
	}
	// Start a goroutine to read and parse messages.
	conn, errConn := file.SyscallConn()
	if errConn != nil {
		return KernelMonitor{}, errConn
	}
	return KernelMonitor{
		f:       file,
		conn:    conn,
		History: make([]KMsg, 0, nMaxEntries),
		Cap:     nMaxEntries,
	}, nil
}

func (p *KernelMonitor) Close() error {
	return p.f.Close()
}

func (p *KernelMonitor) Read(msgCh chan KMsg) error {
	var syscallError error = nil
	err := p.conn.Read(func(fd uintptr) bool {
		for {
			buf := make([]byte, 1024)
			_, err := syscall.Read(int(fd), buf)
			if err != nil {
				syscallError = err
				// EINVAL means buf is not enough, data would be truncated, but still can continue.
				if !errors.Is(err, syscall.EINVAL) {
					return true
				}
			}

			kmsg, err := parseKmsg(string(buf))
			if err != nil {
				fmt.Printf("Error parsing kmsg: %v\n", err)
				continue
			}

			if p.Cap < len(p.History) {
				p.History = p.History[1 : p.Cap-1]
			}
			p.History = append(p.History, kmsg)
			if len(msgCh) < cap(msgCh) { //avoid jamming
				msgCh <- kmsg
			}

		}
	})

	// EAGAIN means no more data, should be treated as normal.
	if syscallError != nil && !errors.Is(syscallError, syscall.EAGAIN) {
		err = syscallError
	}
	return err
}

/*
func MonitorKmsg(msgCh chan KMsg) error {
	file, err := os.OpenFile("/dev/kmsg", syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return fmt.Errorf("error opening /dev/kmsg: %v", err)
	}
	defer file.Close()

	// Start a goroutine to read and parse messages.

	conn, errConn := file.SyscallConn()
	if errConn != nil {
		return errConn
	}

	var syscallError error = nil
	err = conn.Read(func(fd uintptr) bool {
		for {
			buf := make([]byte, 1024)
			_, err := syscall.Read(int(fd), buf)
			if err != nil {
				syscallError = err
				// EINVAL means buf is not enough, data would be truncated, but still can continue.
				if !errors.Is(err, syscall.EINVAL) {
					return true
				}
			}

			kmsg, err := parseKmsg(string(buf))
			if err != nil {
				fmt.Printf("Error parsing kmsg: %v\n", err)
				continue
			}

			msgCh <- kmsg

		}
	})

	// EAGAIN means no more data, should be treated as normal.
	if syscallError != nil && !errors.Is(syscallError, syscall.EAGAIN) {
		err = syscallError
	}



	//close(msgCh)
	return err
}
*/
