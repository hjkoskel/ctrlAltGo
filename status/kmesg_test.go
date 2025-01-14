package status

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKmesgParse(t *testing.T) {
	testEntries := []string{
		"6,1948,111582296275,-;Bluetooth: hci0: Firmware already loaded",
		"4,1949,111582303475,-;Bluetooth: hci0: HCI LE Coded PHY feature bit is set, but its usage is not supported.",
		"6,1950,111727705644,-;usb 3-3: reset full-speed USB device number 11 using xhci_hcd\nSUBSYSTEM=usb\nDEVICE=c189:266",
		"6,1951,111793869816,-;usb 3-3: USB disconnect, device number 11\nSUBSYSTEM=usb\nDEVICE=c189:266",
	}

	for _, entry := range testEntries {
		parsed, errParse := parseKmsg(entry)
		assert.Equal(t, nil, errParse)
		fmt.Printf("parsed %#v\n", parsed)
		//assert.Equal(t, parsed, "")
	}

}
