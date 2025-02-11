/*
Getting settings from textfiles.

Very crude
*/
package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hjkoskel/ctrlaltgo/networking"
	"github.com/hjkoskel/timegopher/timesync"
)

const (
	TZFILENAME   = "tz.txt"
	HOSTFILENAME = "host.txt"

	ETH0FILE_IP      = "eth0.ip"
	ETH0FILE_GATEWAY = "eth0.gw"
	ETH0FILE_NS      = "eth0.ns" //addresses

	NTPSERVERSFILE = "ntp.txt"
)
const (
	DEFAULT_TZ       = "Europe/Helsinki"
	DEFAULT_HOSTNAME = "intraberry"
)

//Load to map... TODO one way to operate or report and maybe change settings?

type AllSettings struct {
	Tz          *time.Location
	Hostname    string
	EthSettings *networking.IpSettings
	NtpSettings timesync.NtpSync
}

func LoadAllSettings(dirname string) (AllSettings, error) {
	result := AllSettings{}
	var err error

	result.Tz, err = GetTz(dirname)
	if err != nil {
		return result, fmt.Errorf("%s", err)
	}

	result.Hostname, err = GetHostname(dirname)
	if err != nil {
		return result, fmt.Errorf("%s", err)
	}
	result.EthSettings, err = GetEthSettings(dirname)
	if err != nil {
		return result, fmt.Errorf("%s", err)
	}

	result.NtpSettings, err = GetNtpSettings(dirname)
	if err != nil {
		return result, fmt.Errorf("%s", err)
	}
	return result, nil
}

func GetTz(dirname string) (*time.Location, error) {
	tzFname := path.Join(dirname, TZFILENAME)
	tzDefault, errTzDefault := time.LoadLocation(DEFAULT_TZ)
	if errTzDefault != nil {
		return tzDefault, errTzDefault
	}
	if !FileExists(tzFname) {
		return tzDefault, nil
	}

	byt, errRead := os.ReadFile(tzFname)
	if errRead != nil {
		return tzDefault, errRead
	}

	tz, errTz := time.LoadLocation(strings.TrimSpace(string(byt)))
	if errTz != nil {
		return tzDefault, errTz
	}
	return tz, nil
}

func GetHostname(dirname string) (string, error) {
	hostFname := path.Join(dirname, HOSTFILENAME)
	if !FileExists(hostFname) {
		return DEFAULT_HOSTNAME, nil
	}
	byt, err := os.ReadFile(hostFname)
	if err != nil {
		return DEFAULT_HOSTNAME, err
	}
	result := strings.TrimSpace(string(byt))
	if len(result) == 0 {
		return DEFAULT_HOSTNAME, nil
	}
	return result, nil
}

func GetEthSettings(dirname string) (*networking.IpSettings, error) {
	fnameIp := path.Join(dirname, ETH0FILE_IP) //IF this is defined then others have to be found also!
	if !FileExists(fnameIp) {
		return nil, nil
	}
	bytIp, errBytIp := os.ReadFile(fnameIp)
	if errBytIp != nil {
		return nil, errBytIp
	}

	result := networking.IpSettings{
		Address: strings.TrimSpace(string(bytIp)),
	}

	fnameGW := path.Join(dirname, ETH0FILE_GATEWAY)
	if FileExists(fnameGW) {
		bytGw, errBytIp := os.ReadFile(fnameGW)
		if errBytIp != nil {
			return nil, errBytIp
		}
		result.Gateway = net.ParseIP(strings.TrimSpace(string(bytGw)))
	}

	fnameEth0 := path.Join(dirname, ETH0FILE_NS)
	if FileExists(fnameEth0) {
		byt, errRead := os.ReadFile(fnameEth0)
		if errRead != nil {
			return nil, errRead
		}
		rows := strings.Split(string(byt), "\n")
		for _, row := range rows {
			s := strings.TrimSpace(row)
			if len(s) == 0 {
				continue
			}
			ipParsed := net.ParseIP(s)
			if ipParsed == nil {
				return &result, fmt.Errorf("error parsing DNS row:%s invalid ip number", row)
			}
			result.DnsServers = append(result.DnsServers, ipParsed)
		}
	} else {
		result.DnsServers = []net.IP{net.ParseIP("8.8.8.8")}
	}

	return &result, nil
}

func GetNtpSettings(dirname string) (timesync.NtpSync, error) {
	fname := path.Join(dirname, NTPSERVERSFILE)
	if !FileExists(fname) {
		return timesync.GetDefaultFinnishNTP(), nil
	}

	byt, errRead := os.ReadFile(fname)
	if errRead != nil {
		return timesync.GetDefaultFinnishNTP(), errRead
	}

	rows := strings.Split(string(byt), "\n")

	srvlist := []string{}
	for _, row := range rows {
		s := strings.TrimSpace(row)
		if !strings.HasPrefix(s, "#") && 0 < len(s) {
			rows = append(rows, s)
		}
	}

	return timesync.NtpSync{
		Servers:      srvlist,
		QueryTimeout: time.Second * 30,
	}, nil
}
