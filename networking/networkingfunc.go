/*
Internal networking library
General util functions for networking

TODO mockuping?
*/

package networking

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type IpSettings struct {
	Gateway    net.IP //Goes to route
	Address    string
	DnsServers []net.IP
	LeaseTime  time.Duration //Stays same
	Expire     time.Time     //If leased from DHCP
}

func (a IpSettings) String() string {
	if len(a.Address) == 0 {
		return "IPnotDefined"
	}

	dnslist := make([]string, len(a.DnsServers))
	for i, d := range a.DnsServers {
		dnslist[i] = d.String()
	}
	return fmt.Sprintf("IP:%s GW:%s DNS:[%s] lease:%s", a.Address, a.Gateway, strings.Join(dnslist, ","), a.LeaseTime)
}

func (p *IpSettings) ApplyToInterface(interfacename string, priority int) error {
	errAddr := SetAddress(interfacename, p.Address)
	if errAddr != nil {
		return fmt.Errorf("error setting address %s on %s err:%s", p.Address, interfacename, errAddr)
	}

	errRoute := SetRoute(interfacename, p.Gateway, priority)
	if errRoute != nil {
		return fmt.Errorf("error setting route on %s gw:%s priority:%v err:%s", interfacename, p.Gateway, priority, errRoute)
	}

	resolvconf, errResolvRead := ReadResolvConf(RESOLVCONFFILE)
	if errResolvRead != nil {
		return fmt.Errorf("error reading %s err:%s", RESOLVCONFFILE, errResolvRead)
	}

	errNsAdd := resolvconf.AddNameservers(p.DnsServers)
	if errNsAdd != nil {
		return fmt.Errorf("error adding nameservers %s", errNsAdd)
	}

	errSave := resolvconf.Save(RESOLVCONFFILE)
	if errSave != nil {
		return fmt.Errorf("error saving %s", RESOLVCONFFILE)
	}
	return nil
}

func GetDHCP(hostname string, interfacename string) (IpSettings, error) {
	interf, errInterf := net.InterfaceByName(interfacename)
	if errInterf != nil {
		return IpSettings{}, fmt.Errorf("interface %s err:%s", interfacename, errInterf)
	}

	lease, errLease := GetDhcpLease(interf, hostname)
	if errLease != nil {
		return IpSettings{}, fmt.Errorf("DHCP err %s", errLease)
	}
	result := IpSettings{
		Address:    lease.IP.String() + "/24",
		DnsServers: lease.DNS,
		Gateway:    lease.Router,
		LeaseTime:  lease.RenewalTime,
		Expire:     time.Now().Add(lease.RenewalTime),
	}

	if len(lease.Netmask) > 0 {
		a := net.IPNet{IP: lease.IP, Mask: lease.Netmask}
		result.Address = a.String()
	}
	return result, nil
}

func Link(interfacename string) (bool, error) {
	link, errLink := netlink.LinkByName(interfacename)
	if errLink != nil {
		return false, errLink
	}
	return link.Attrs().OperState == netlink.OperUp, nil
}

func Carrier(interfacename string) (bool, error) {
	b, err := os.ReadFile(filepath.Join("/sys/class/net", interfacename, "carrier"))
	return strings.TrimSpace(string(b)) == "1", err
}

// WaitCarrier, helper function for polling carrier
func WaitCarrier(interfacename string, timeout time.Duration, interval time.Duration) error {
	t0 := time.Now()
	for time.Since(t0) < timeout {
		have, err := Carrier(interfacename)
		if err != nil {
			return err
		}
		if have {
			return nil //GOOD, got
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("timeout")
}

func ListInterfaceNames() ([]string, error) {
	ifaces, errList := net.Interfaces()
	if errList != nil {
		return nil, errList
	}
	result := make([]string, len(ifaces))
	for i, a := range ifaces {
		result[i] = a.Name
	}
	return result, nil
}

func HaveInterface(interfacename string) (bool, error) {
	lst, err := ListInterfaceNames()
	if err != nil {
		return false, err
	}
	for _, s := range lst {
		if s == interfacename {
			return true, nil
		}
	}
	return false, nil
}

func WaitInterface(interfacename string, timeout time.Duration, interval time.Duration) error {
	t0 := time.Now()
	for time.Since(t0) < timeout {
		have, err := HaveInterface(interfacename)
		if err != nil {
			return err
		}
		if have {
			return nil //GOOD, got
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("timeout")
}

func SetLinkUp(interfacename string, up bool) error {
	link, errLink := netlink.LinkByName(interfacename)
	if errLink != nil {
		return errLink
	}

	opState := link.Attrs().OperState

	wanted := netlink.OperDown
	if up {
		wanted = netlink.OperUp
	}

	handle, errHandle := netlink.NewHandle(netlink.FAMILY_V4)
	if errHandle != nil {
		return errHandle
	}

	if opState != netlink.LinkOperState(wanted) { //Rise only if link is
		var errSet error
		if up {
			errSet = handle.LinkSetUp(link)
		} else {
			errSet = handle.LinkSetDown(link)
		}
		if errSet != nil {
			return fmt.Errorf("error setting link %s", errSet)
		}
	}
	return nil
}

func SetAddress(interfName string, addrWithMask string) error {
	if !strings.Contains(addrWithMask, "/") {
		addrWithMask += "/24" //Default mask  255.255.255.0
	}
	addr, parseErr := netlink.ParseAddr(addrWithMask)
	if parseErr != nil {
		return fmt.Errorf("error parsing %s err=%s", addrWithMask, parseErr)
	}
	link, errLink := netlink.LinkByName(interfName)
	if errLink != nil {
		return errLink
	}
	errAddrReplace := netlink.AddrReplace(link, addr)
	if errAddrReplace != nil {
		return fmt.Errorf("AddrReplace: %v", errAddrReplace)
	}
	return nil
}

func GetInterfaceByName(interfacename string, timeout time.Duration) (*net.Interface, error) {
	t0 := time.Now()
	for time.Since(t0) < timeout {
		interf, errInterf := net.InterfaceByName(interfacename)
		if errInterf == nil {
			return interf, nil
		}
	}
	return nil, fmt.Errorf("timeout")
}

func SetRoute(interfacename string, gatewayIp net.IP, priority int) error {
	link, errLink := netlink.LinkByName(interfacename)
	if errLink != nil {
		return errLink
	}

	return netlink.RouteReplace(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       defaultDst,
		Gw:        gatewayIp,
		Priority:  priority,
	})
}
