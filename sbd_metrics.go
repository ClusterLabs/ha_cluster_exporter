package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func readSdbFile() ([]byte, error) {
	sbdConfFile, err := os.Open("/etc/sysconfig/sbd")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Could not open sbd config file %s", err)
	}

	defer sbdConfFile.Close()
	sbdConfigRaw, err := ioutil.ReadAll(sbdConfFile)

	if err != nil {
		return nil, fmt.Errorf("[ERROR] Could not read sbd config file %s", err)
	}
	return sbdConfigRaw, nil
}

// return a list of sbd devices that we get from config
func getSbdDevices(sbdConfigRaw []byte) []string {
	// in config it can be both SBD_DEVICE="/dev/foo" or SBD_DEVICE=/dev/foo;/dev/bro
	wordOnly := regexp.MustCompile("SBD_DEVICE=\"?[a-zA-Z-/;]+\"?")
	sbdDevicesConfig := wordOnly.FindString(string(sbdConfigRaw))
	// remove the SBD_DEVICE
	sbdArray := strings.Split(sbdDevicesConfig, "SBD_DEVICE=")[1]
	// make a list of devices by ; seperators and remove double quotes if present
	sbdDevices := strings.Split(strings.Trim(sbdArray, "\""), ";")

	return sbdDevices
}

// this function take a list of sbd devices and return
// a  map of devices with the status, true is healthy , false isn't
func setSbdDeviceHealth(sbdDevices []string) map[string]bool {
	sbdStatus := make(map[string]bool)

	for _, sbdDev := range sbdDevices {
		_, err := exec.Command("sbd", "-d", sbdDev, "dump").Output()

		// in case of error the device is not healthy
		if err != nil {
			sbdStatus[sbdDev] = false
		} else {
			sbdStatus[sbdDev] = true
		}
	}
	return sbdStatus
}
