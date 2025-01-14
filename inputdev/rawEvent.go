package inputdev

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

/*
https://www.kernel.org/doc/Documentation/input/input.txt

/usr/include/linux/input-event-codes.h
*/

type RawEventType uint16

const (
	EV_SYN       RawEventType = 0x00
	EV_KEY       RawEventType = 0x01
	EV_REL       RawEventType = 0x02
	EV_ABS       RawEventType = 0x03
	EV_MSC       RawEventType = 0x04
	EV_SW        RawEventType = 0x05
	EV_LED       RawEventType = 0x11
	EV_SND       RawEventType = 0x12
	EV_REP       RawEventType = 0x14
	EV_FF        RawEventType = 0x15
	EV_PWR       RawEventType = 0x16
	EV_FF_STATUS RawEventType = 0x17
	EV_MAX       RawEventType = 0x1f
)

func (a RawEventType) String() string {
	if EV_MAX < a {
		return "UNK"
	}
	return []string{"SYN", "KEY", "REL", "ABS", "MSC", "SW", "LED", "SND", "REP", "FF", "PWR", "FF_STATUS", "MAX"}[a]
}

type KeyCode uint16

type RawInputEvent struct {
	Timestamp time.Time
	Type      RawEventType // /usr/include/linux/input-event-codes.h
	Code      uint16       // event code related to the event type
	Value     uint32
}

func (a RawInputEvent) String() string {
	return fmt.Sprintf("%s %s code:0x%02X val:%v", a.Timestamp, a.Type, a.Code, a.Value)
}

func ReadInputEvent(f *os.File) (RawInputEvent, error) {
	b := make([]byte, 24)

	_, errRead := f.Read(b)
	if errRead != nil {
		return RawInputEvent{}, errRead
	}
	sec := binary.LittleEndian.Uint64(b[0:8])
	usec := binary.LittleEndian.Uint64(b[8:16])
	t := time.Unix(int64(sec), int64(usec))

	return RawInputEvent{
		Timestamp: t,
		Type:      RawEventType(binary.LittleEndian.Uint16(b[16:18])),
		Code:      binary.LittleEndian.Uint16(b[18:20]),
		Value:     binary.LittleEndian.Uint32(b[20:])}, nil
}
