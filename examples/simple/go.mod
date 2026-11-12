module simple

go 1.25.0

require github.com/hjkoskel/ctrlaltgo/initializing v0.0.2


require (
	ctrlaltgo v0.0.0
	networking v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/gopacket v1.1.19 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mdlayher/packet v1.1.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/rtr7/dhcp4 v0.0.0-20220302171438-18c84d089b46 // indirect
	github.com/vishvananda/netlink v1.3.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace ctrlaltgo => ../../

replace networking => ../../networking
