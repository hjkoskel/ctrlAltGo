/*
List block devices
*/
package initializing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SizeBytes int64 //TODO to some generic lib?

func (a SizeBytes) StringDecimal() string {
	//https://en.wikipedia.org/wiki/Byte#Multiple-byte_units
	if a < 1000 {
		return fmt.Sprintf("%d", int(a))
	}
	if a < 1000*1000 {
		return fmt.Sprintf("%.2fkB", float64(a)/1000)
	}
	if a < 1000*1000*1000 {
		return fmt.Sprintf("%.2fMB", float64(a)/(1000*1000))
	}
	if a < 1000*1000*1000*1000 {
		return fmt.Sprintf("%.2fGB", float64(a)/(1000*1000*1000))
	}
	return fmt.Sprintf("%.2fTB", float64(a)/(1000*1000*1000*1000))
}

func (a SizeBytes) String() string {
	//https://en.wikipedia.org/wiki/Byte#Multiple-byte_units
	if a < 1024 {
		return fmt.Sprintf("%d", int(a))
	}
	if a < 1024*1024 {
		return fmt.Sprintf("%.2fkiB", float64(a)/1024)
	}
	if a < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMiB", float64(a)/(1024*1024))
	}
	if a < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fGiB", float64(a)/(1024*1024*1024))
	}
	return fmt.Sprintf("%.2fTiB", float64(a)/(1024*1024*1024*1024))
}

// BlockDevice represents information about a block device
type BlockDevice struct {
	Name       string
	SizeBytes  SizeBytes
	Vendor     string
	Model      string
	Partitions []Partition
	IsUSB      bool
}

func (a BlockDevice) String() string {
	var sb strings.Builder
	if a.IsUSB {
		sb.WriteString("USB:")
	}
	sb.WriteString(fmt.Sprintf("%s: %s model:%s", a.Name, a.SizeBytes, a.Model))
	for i, part := range a.Partitions {
		sb.WriteString(fmt.Sprintf("  %d:%s\n", i, part))
	}
	return sb.String()
}

type BlockDevices []BlockDevice

func (p *BlockDevices) String() string {
	var sb strings.Builder
	for _, dev := range *p {
		sb.WriteString(dev.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

type Partition struct {
	Name   string
	SizeGB int64
	UUID   string
	FSType string
}

func (a Partition) String() string {
	return fmt.Sprintf("%s size:%v fstype:%s UUID:%s", a.Name, a.SizeGB, a.FSType, a.UUID)
}

// getPartitionInfo retrieves information about a partition
func getPartitionInfo(partitionPath string) (Partition, error) {
	par := Partition{
		Name: filepath.Base(partitionPath),
	}

	// Read partition size in 512-byte sectors and convert to GB
	sizeSectorsInt, errSize := ReadIntfile(filepath.Join(partitionPath, "size"))
	if errSize != nil {
		return par, errSize
	}
	par.SizeGB = (sizeSectorsInt * 512) / (1024 * 1024 * 1024)

	// Read UUID and filesystem type (if available)
	byt, errUuid := os.ReadFile(filepath.Join(partitionPath, "uuid"))
	if errUuid != nil {
		return par, errUuid
	}
	par.UUID = string(byt)

	bytUevent, errUevent := os.ReadFile(filepath.Join(partitionPath, "uevent"))
	if errUevent != nil {
		return par, errUevent
	}
	par.FSType = string(bytUevent)
	if strings.Contains(par.FSType, "ID_FS_TYPE=") {
		par.FSType = strings.Split(par.FSType, "ID_FS_TYPE=")[1]
		par.FSType = strings.Split(par.FSType, "\n")[0]
	}

	return par, nil
}

func GetBlockDeviceNames() ([]string, error) {
	// List all block devices in /sys/block
	dirArr, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("failed to read /sys/block: %v", err)
	}
	result := []string{}
	for _, d := range dirArr {
		name := d.Name()
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "nbd") {
			continue
		}
		result = append(result, name)
	}
	return result, nil
}

const UNKNOWSTRING string = "unknown"

// getBlockDevices parses information from /sys/block
func GetBlockDevices() (BlockDevices, error) {
	var devices []BlockDevice

	blkNames, errBlkNames := GetBlockDeviceNames()
	if errBlkNames != nil {
		return nil, errBlkNames
	}

	for _, deviceName := range blkNames {
		devicePath := filepath.Join("/sys/block", deviceName)
		device := BlockDevice{
			Name: deviceName,
		}

		// Read device size in 512-byte sectors and convert to GB
		sizeSectorsInt, errSizeSectorsInt := ReadIntfile(filepath.Join(devicePath, "size"))
		if errSizeSectorsInt != nil {
			return devices, errSizeSectorsInt
		}
		device.SizeBytes = SizeBytes(sizeSectorsInt * 512)

		byt, errByt := os.ReadFile(filepath.Join(devicePath, "device", "vendor"))

		if errByt != nil {
			device.Vendor = UNKNOWSTRING
		} else {
			device.Vendor = string(byt)
		}

		byt, errByt = os.ReadFile(filepath.Join(devicePath, "device", "model"))
		if errByt != nil {
			device.Model = UNKNOWSTRING
		} else {
			device.Model = string(byt)
		}

		// Check if the device is a USB drive
		deviceLink, errReadLink := os.Readlink(filepath.Join("/sys", "class", "block", deviceName))
		if errReadLink != nil {
			return devices, errReadLink
		}
		//fmt.Printf("DEVICE LINK IS %s\n", deviceLink)
		device.IsUSB = strings.Contains(deviceLink, "/usb")

		// List partitions (if any)
		//var partitions []string
		partitionEntries, _ := os.ReadDir(devicePath)

		/*for i, partitionEntry := range partitionEntries {
			fmt.Printf("DEBUG: partition entry %v:%s\n", i, partitionEntry.Name())
		}*/

		for _, partitionEntry := range partitionEntries {
			//fmt.Printf("DOES %s have prefix %s  but not same\n", partitionEntry.Name(), deviceName)
			if strings.HasPrefix(partitionEntry.Name(), deviceName) && partitionEntry.Name() != deviceName {
				//partitions = append(partitions, partitionEntry.Name())
				partInfo, _ := getPartitionInfo(path.Join(devicePath, partitionEntry.Name()))
				device.Partitions = append(device.Partitions, partInfo)
			}
		}
		// Create BlockDevice struct

		devices = append(devices, device)
	}

	return devices, nil
}
