package inputdev

import (
	"fmt"
	"os"
	"strings"
)

/*
https://www.kernel.org/doc/html/latest/input/event-codes.html
https://unix.stackexchange.com/questions/74903/explain-ev-in-proc-bus-input-devices-data

const(
	#define EV_SYN                  0x00
#define EV_KEY                  0x01
#define EV_REL                  0x02
#define EV_ABS                  0x03
#define EV_MSC                  0x04
#define EV_SW                   0x05
#define EV_LED                  0x11
#define EV_SND                  0x12
#define EV_REP                  0x14
#define EV_FF                   0x15
#define EV_PWR                  0x16
#define EV_FF_STATUS            0x17
#define EV_MAX                  0x1f
*/

/*
type InputDeviceBitmap struct{
	Prop string //PROP => device properties and quirks.
	Ev string //EV   => types of events supported by the device.
	KEY string //  => keys/buttons this device has.
	MSC string // => miscellaneous events supported by the device.
	LED scrint // => leds present on the device.
}
*/

// InputDevice represents an input device from /proc/bus/input/devices
type InputDevice struct {
	Handlers []string
	Name     string
	Phys     string
	Sysfs    string
	Bus      string
	Vendor   string
	Product  string
	Version  string
	//Bitmaps
	Bitmaps map[string]string //HACK, just extranct what code says
}

/*
B => bitmaps

	PROP => device properties and quirks.
	EV   => types of events supported by the device.
	KEY  => keys/buttons this device has.
	MSC  => miscellaneous events supported by the device.
	LED  => leds present on the device.


	https://github.com/torvalds/linux/blob/master/include/uapi/linux/input-event-codes.h

	https://unix.stackexchange.com/questions/74903/explain-ev-in-proc-bus-input-devices-data
*/
type InputDeviceArr []InputDevice

func (p *InputDevice) GetDevName() string {
	for _, hand := range p.Handlers {
		hand = strings.Replace(hand, "Handlers=", "", 1)
		//fmt.Printf("TESTING %s\n", hand)
		if strings.HasPrefix(hand, "event") {
			return fmt.Sprintf("/dev/input/%s", hand)
		}
	}
	return ""
}

func (p *InputDeviceArr) GetDefaultKeyboards() InputDeviceArr {
	result := []InputDevice{}
	for _, itm := range *p {
		evField, hazEV := itm.Bitmaps["EV"]
		if !hazEV {
			continue
		}
		if evField == "120013" {
			result = append(result, itm)
			return result
		}
	}
	return result
}

func (p *InputDeviceArr) GetMices() InputDeviceArr {
	result := []InputDevice{}
	for _, dev := range *p {
		for _, hand := range dev.Handlers {
			if strings.HasPrefix(hand, "mouse") {
				result = append(result, dev)
			}
		}
	}
	return result
}

// ParseDevicesFile parses the /proc/bus/input/devices file and returns a slice of InputDevice
func ParseDevicesFile(filePath string) (InputDeviceArr, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var devices []InputDevice
	currentDevice := InputDevice{Bitmaps: make(map[string]string)}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// End of a device block
			if currentDevice.Name != "" {
				devices = append(devices, currentDevice)
			}
			currentDevice = InputDevice{Bitmaps: make(map[string]string)}
			continue
		}

		switch {
		case strings.HasPrefix(line, "N:"):
			currentDevice.Name = parseField(line)
		case strings.HasPrefix(line, "P:"):
			currentDevice.Phys = parseField(line)
		case strings.HasPrefix(line, "S:"):
			currentDevice.Sysfs = parseField(line)
		case strings.HasPrefix(line, "H:"):
			handlers := strings.Fields(line[len("H:"):])
			currentDevice.Handlers = append(currentDevice.Handlers, handlers...)
		case strings.HasPrefix(line, "B:"):
			kv := strings.Split(strings.Replace(line, "B:", "", 1), "=")
			currentDevice.Bitmaps[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		case strings.HasPrefix(line, "I:"):
			fields := strings.Fields(line[len("I:"):])
			for _, field := range fields {
				kv := strings.SplitN(field, "=", 2)
				if len(kv) == 2 {
					switch kv[0] {
					case "Bus":
						currentDevice.Bus = kv[1]
					case "Vendor":
						currentDevice.Vendor = kv[1]
					case "Product":
						currentDevice.Product = kv[1]
					case "Version":
						currentDevice.Version = kv[1]
					}
				}
			}
		}
	}

	// Add the last device block if it exists
	if currentDevice.Name != "" {
		devices = append(devices, currentDevice)
	}

	return devices, nil
}

// parseField extracts the value from lines like "N: Name=\"Some Name\""
func parseField(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) == 2 {
		return strings.Trim(parts[1], "\"")
	}
	return ""
}
