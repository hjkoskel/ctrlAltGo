module runllamafile

go 1.23.2

replace initializing => ../../initializing

replace ctrlaltgo => ../../

require (
	github.com/hjkoskel/ctrlaltgo v0.0.1
	github.com/hjkoskel/ctrlaltgo/initializing v0.0.0-20250222161100-80d17713b5bd
	github.com/hjkoskel/ctrlaltgo/networking v0.0.0-20250222161100-80d17713b5bd
)

require (
	github.com/google/gopacket v1.1.19 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mdlayher/packet v1.1.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/rtr7/dhcp4 v0.0.0-20220302171438-18c84d089b46 // indirect
	github.com/vishvananda/netlink v1.3.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)

replace networking => ../../networking
