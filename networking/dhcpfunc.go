/*
copied from gokrazy
*/
package networking

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/mdlayher/packet"
	"github.com/rtr7/dhcp4"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type client struct {
	hostname     string
	hardwareAddr net.HardwareAddr
	generateXID  func() uint32
	conn         net.PacketConn
}

func (c *client) packet(xid uint32, opts []layers.DHCPOption) *layers.DHCPv4 {
	return &layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  uint8(len(layers.EthernetBroadcast)),
		HardwareOpts: 0, // clients set this to zero (used by relay agents)
		Xid:          xid,
		Secs:         0, // TODO: fill in?
		Flags:        0, // we can receive IP packets via unicast
		ClientHWAddr: c.hardwareAddr,
		ServerName:   nil,
		File:         nil,
		Options:      opts,
	}
}

func (c *client) discover() (*layers.DHCPv4, error) {
	discover := c.packet(c.generateXID(), []layers.DHCPOption{
		dhcp4.MessageTypeOpt(layers.DHCPMsgTypeDiscover),
		dhcp4.HostnameOpt(c.hostname),
		dhcp4.ClientIDOpt(layers.LinkTypeEthernet, c.hardwareAddr),
		dhcp4.ParamsRequestOpt(
			layers.DHCPOptDNS,
			layers.DHCPOptRouter,
			layers.DHCPOptSubnetMask,
			layers.DHCPOptDomainName),
	})
	if err := dhcp4.Write(c.conn, discover); err != nil {
		return nil, err
	}

	// Look for DHCPOFFER packet (described in RFC2131 4.3.1):
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		offer, err := dhcp4.Read(c.conn)
		if err != nil {
			return nil, err
		}
		if offer == nil {
			continue // not a DHCPv4 packet
		}
		if offer.Xid != discover.Xid {
			continue // broadcast reply for different DHCP transaction
		}
		if !dhcp4.HasMessageType(offer.Options, layers.DHCPMsgTypeOffer) {
			continue
		}
		return offer, nil
	}
}

func (c *client) request(last *layers.DHCPv4) (*layers.DHCPv4, error) {
	// Build a DHCPREQUEST packet:
	request := c.packet(last.Xid, append([]layers.DHCPOption{
		dhcp4.MessageTypeOpt(layers.DHCPMsgTypeRequest),
		dhcp4.RequestIPOpt(last.YourClientIP),
		dhcp4.HostnameOpt(c.hostname),
		dhcp4.ClientIDOpt(layers.LinkTypeEthernet, c.hardwareAddr),
		dhcp4.ParamsRequestOpt(
			layers.DHCPOptDNS,
			layers.DHCPOptRouter,
			layers.DHCPOptSubnetMask,
			layers.DHCPOptDomainName),
	}, dhcp4.ServerID(last.Options)...))
	if err := dhcp4.Write(c.conn, request); err != nil {
		return nil, err
	}

	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		// Look for DHCPACK packet (described in RFC2131 4.3.1):
		ack, err := dhcp4.Read(c.conn)
		if err != nil {
			return nil, err
		}
		if ack == nil {
			continue // not a DHCPv4 packet
		}
		if ack.Xid != request.Xid {
			continue // broadcast reply for different DHCP transaction
		}
		if !dhcp4.HasMessageType(ack.Options, layers.DHCPMsgTypeAck) {
			continue
		}
		return ack, nil
	}
}

func GetDhcpLease(ifi *net.Interface, hostname string) (*dhcp4.Lease, error) {
	conn, err := packet.Listen(ifi, packet.Datagram, unix.ETH_P_IP, nil)
	if err != nil {
		log.Fatal(err)
	}

	c := &client{
		hostname:     hostname,
		hardwareAddr: ifi.HardwareAddr,
		generateXID:  dhcp4.XIDGenerator(ifi.HardwareAddr),
		conn:         conn,
	}

	offer, errDiscover := c.discover()
	if err != nil {
		return nil, errDiscover
	}
	if offer == nil {
		return nil, fmt.Errorf("dhcp discover returns nil offer")
	}
	//TODO is this needed?
	var errRequest error
	offer, errRequest = c.request(offer)
	if errRequest != nil {
		return nil, errRequest
	}
	result := dhcp4.LeaseFromACK(offer)

	return &result, nil
}

func changeRoutePriority(nl *netlink.Handle, l netlink.Link, priority int) error {
	routes, err := nl.RouteList(l, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("netlink.RouteList: %v", err)
	}
	for _, route := range routes {
		if route.Priority == priority {
			continue // no change necessary
		}
		newRoute := route // copy
		log.Printf("adjusting route [dst=%v src=%v gw=%v] priority to %d", route.Dst, route.Src, route.Gw, priority)
		newRoute.Flags = 0 // prevent "invalid argument" error
		newRoute.Priority = priority
		if err := nl.RouteReplace(&newRoute); err != nil {
			return fmt.Errorf("RouteReplace: %v", err)
		}
		if err := nl.RouteDel(&route); err != nil {
			return fmt.Errorf("RouteDel: %v", err)
		}
	}
	return nil
}

func priorityFromName(ifname string) int {
	if strings.HasPrefix(ifname, "eth") {
		return 1
	}
	return 5 // wlan0 and others
}

var defaultDst = func() *net.IPNet {
	_, net, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		log.Fatal(err)
	}
	return net
}()

// TODO TARVIIKO?
func applyLease(nl *netlink.Handle, ifname string, lease dhcp4.Lease, extraRoutePriority int) error {
	// Log the received DHCPACK packet:
	addrstr := lease.IP.String() + "/24"
	if len(lease.Netmask) > 0 {
		ipnet := net.IPNet{
			IP:   lease.IP,
			Mask: lease.Netmask,
		}
		addrstr = ipnet.String()
	}

	l, err := nl.LinkByName(ifname)
	if err != nil {
		return fmt.Errorf("LinkByName: %v", err)
	}

	// Apply the received settings:
	addr, err := netlink.ParseAddr(addrstr)
	if err != nil {
		return err
	}
	if err := nl.AddrReplace(l, addr); err != nil {
		return fmt.Errorf("AddrReplace: %v", err)
	}

	if l.Attrs().OperState != netlink.OperUp {
		if err := nl.LinkSetUp(l); err != nil {
			return fmt.Errorf("LinkSetUp: %v", err)
		}
	}

	// Adjust the priority of the network routes on this interface; the kernel
	// adds at least one based on the configured address.
	if err := changeRoutePriority(nl, l, priorityFromName(ifname)+extraRoutePriority); err != nil {
		return fmt.Errorf("changeRoutePriority: %v", err)
	}
	fmt.Printf("route priorietes are set (\n")
	if r := lease.Router; len(r) > 0 {
		fmt.Printf("Replacing route\n")
		err = nl.RouteReplace(&netlink.Route{
			LinkIndex: l.Attrs().Index,
			Dst:       defaultDst,
			Gw:        r,
			Priority:  priorityFromName(ifname) + extraRoutePriority,
		})
		if err != nil {
			return fmt.Errorf("RouteReplace: %v", err)
		}
	}

	fmt.Printf("lease had DNS servers %#v\n", lease.DNS)

	return nil
}
