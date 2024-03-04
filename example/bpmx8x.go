package main

import (
	"encoding/binary"
	"log"
	"os"

	"zappem.net/pub/io/i2c"
)

// Package BPM pressure sensor device.
//
// Details:
//
//   https://community.bosch-sensortec.com/t5/Knowledge-base/BMP-series-pressure-sensor-design-guide/ta-p/7103

// ids for the two supported devices
var ids = map[byte]string{
	0x58: "BMP280",
	0x50: "BMP388",
}

// addresses supported by the BPM devices
var addrs = []uint{0x76, 0x77}

func main() {
	found := false
	for _, i := range addrs {
		c, err := i2c.NewConn(i2c.BusFile(1), i, false, binary.LittleEndian)
		if err != nil {
			continue
		}
		// Read BPM device ID.
		val, err := c.Reg(0x00)
		if err != nil {
			c.Close()
			log.Printf("failed to read register 0 of device @ %02xh: %v", i, err)
			continue
		}
		dev, ok := ids[val]
		if !ok {
			c.Close()
			log.Printf("unrecognized device @ %02xh ID:%02xh", i, val)
			continue
		}
		log.Printf("Found device %q @ %02xh", dev, i)
		found = true
		c.Close()
	}
	if !found {
		log.Print("BPM sensor not found.")
		os.Exit(1)
	}
}
