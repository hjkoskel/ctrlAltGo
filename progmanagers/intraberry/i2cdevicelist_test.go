package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestI2Cdb(t *testing.T) {
	byt, errRead := os.ReadFile("i2cdevices.json")
	assert.Equal(t, nil, errRead)

	var db I2CDeviceDatabase
	errParse := json.Unmarshal(byt, &db)
	assert.Equal(t, nil, errParse)

	lst := db.DevicesWithAddress(118)

	t.Log(fmt.Sprintf("PartNumbers: %s\n", strings.Join(lst.PartNumbers(), ", ")))
	assert.Equal(t, len(lst), 14)

	//Just check that format is not changed suddenly
	q := db.GetPartNumber("BME280")

	assert.Equal(t, q.PartNumber, "BME280")
	assert.Equal(t, q.FriendlyName, "Temp/Barometric/Humidity")
	assert.Equal(t, q.DataSheet, "https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280_DS002.pdf")
	assert.Equal(t, q.Adafruit, "https://www.adafruit.com/product/2652")
	assert.Equal(t, q.Obsolete, false)
	assert.Equal(t, q.Manufacturer, "Bosch")
	assert.Equal(t, q.Is5V, false)

	assert.Equal(t, q.Is3V, true)
	assert.Equal(t, q.IsSPI, true)
	assert.Equal(t, q.ReleaseDate, 2016)
	assert.Equal(t, q.ScannedDrivers, true)
	assert.Equal(t, q.Drivers, []I2CLinkTitle{}) //TODO check with device with listed drivers
	assert.Equal(t, q.Addresses, []int{119, 118})

}

/* TODO MAKE SOMEKIND OF GO GENERATE THING FOR GENERATING MINIMIZED I2C DEVICE DATABASE
func TestI2CwriteMinimal(t *testing.T) {
	byt, errRead := os.ReadFile("i2cdevices.json")
	assert.Equal(t, nil, errRead)

	var db I2CDeviceDatabase
	errParse := json.Unmarshal(byt, &db)
	assert.Equal(t, nil, errParse)

	db.Minimize()

	resultByt, _ := json.MarshalIndent(db, "", " ")
	os.WriteFile("i2cdevices_min.json", resultByt, 0666)
}
*/
