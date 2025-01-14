module simplesrv

go 1.23.2

require initializing v0.0.0

replace initializing => ../../initializing

require networking v0.0.0

replace networking => ../../networking

require ctrlaltgo v0.0.0

require (
	github.com/beevik/ntp v1.4.3 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/hjkoskel/fixregsto v0.1.0-beta.2 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mdlayher/packet v1.1.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/rtr7/dhcp4 v0.0.0-20220302171438-18c84d089b46 // indirect
	github.com/vishvananda/netlink v1.3.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace ctrlaltgo => ../../

require (
	github.com/hjkoskel/timegopher v0.0.2
	github.com/hjkoskel/timegopher/timesync v0.0.0-20250112124711-4a8804e861a6
)
