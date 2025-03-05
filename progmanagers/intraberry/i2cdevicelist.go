/*
Downloaded json from
https://i2cdevices.org/devices

# Parsing those for GUI

Use this functionality for full database or just for decoding I2C device names (110k is still memory consumed...)
*/
package main

type I2CLinkTitle struct {
	Title string
	Link  string
}

type I2CDeviceInfo struct {
	PartNumber   string `json:"part_number,omitempty"`
	FriendlyName string `json:"friendly_name,omitempty"`
	DataSheet    string `json:"datasheet,omitempty"`
	Adafruit     string `json:"adafruit,omitempty"`
	Obsolete     bool   `json:"obsolete,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Is5V         bool   `json:"is_5v,omitempty"`
	Is3V         bool   `json:"is_3v,omitempty"`
	IsSPI        bool   `json:"is_spi,omitempty"`
	ReleaseDate  int    `json:"release_date,omitempty"`

	ScannedDrivers bool           `json:"scanned_drivers,omitempty"`
	Drivers        []I2CLinkTitle `json:",drivers,omitempty"`
	Addresses      []int          `json:",addresses,omitempty"`
}

func (p *I2CDeviceInfo) Minimize() {
	*p = I2CDeviceInfo{PartNumber: p.PartNumber, Addresses: p.Addresses}
}

type I2CDeviceDatabase []I2CDeviceInfo

func (p *I2CDeviceDatabase) Minimize() {
	for i, _ := range *p {
		(*p)[i].Minimize()
	}
}

func (p *I2CDeviceDatabase) DevicesWithAddress(addrToBeFound int) I2CDeviceDatabase {
	result := []I2CDeviceInfo{}

	for _, device := range *p {
		for _, a := range device.Addresses {
			if a == addrToBeFound {
				result = append(result, device)
				break
			}
		}
	}
	return result
}

// Get namelist... usual use is after filtering with DevicesWithAddress
func (p *I2CDeviceDatabase) PartNumbers() []string {
	result := []string{}
	for _, device := range *p {
		result = append(result, device.PartNumber)
	}
	return result
}

func (p *I2CDeviceDatabase) GetPartNumber(part_number string) I2CDeviceInfo {
	for _, device := range *p {
		if device.PartNumber == part_number {
			return device
		}
	}
	return I2CDeviceInfo{}
}
