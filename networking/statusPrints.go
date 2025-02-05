package networking

import (
	"fmt"
	"strings"
	"time"
)

// TODO MOVE THIS!
func PrintoutNetInterface(interfacename string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: ", interfacename))
	haveLink, errLink := Link(interfacename)
	if errLink != nil {
		sb.WriteString(fmt.Sprintf("ERR %s", errLink))
		return sb.String()
	}

	haveCarrier, errCarrier := Carrier(interfacename)
	if errCarrier != nil {
		sb.WriteString(fmt.Sprintf("ERR %s", errCarrier))
		return sb.String()
	}

	if !haveCarrier {
		sb.WriteString("No carr ")
		if !haveLink {
			sb.WriteString("NoLink ")
		}
		return sb.String()
	}
	if !haveLink {
		sb.WriteString("NoLink ")
		return sb.String()
	}

	//IP number, MAC address and what addresses
	interf, errInterf := GetInterfaceByName(interfacename, time.Second)
	if errInterf != nil {
		sb.WriteString(fmt.Sprintf("interfErr:%s", errInterf))
		return sb.String()
	}
	addrList, errListAddr := interf.Addrs()
	if errListAddr != nil {
		sb.WriteString(fmt.Sprintf("errAddr:%s", errListAddr))
		return sb.String()
	}
	for i, addr := range addrList {
		sb.WriteString(fmt.Sprintf("%v:%s ", i, addr.String()))
	}
	sb.WriteString(fmt.Sprintf("\n  %s\n\n", interf.HardwareAddr))
	return sb.String()
}
