package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// return a list of sbd devices
func getSbdDevices() error {
	// get ringStatus

	sbdConfFile, err := os.Open("/etc/sysconfig/sbd")
	if err != nil {
		return fmt.Errorf("[ERROR] Could not open sbd config file %s", err)
	}

	defer sbdConfFile.Close()
	sbdConfigRaw, err := ioutil.ReadAll(sbdConfFile)

	if err != nil {
		return fmt.Errorf("[ERROR] Could not read sbd config file %s", err)
	}
	sbdDevicesRaw := strings.SplitAfter(string(sbdConfigRaw), "SBD_DEVICE=")
	sbdDevices := strings.Trim(sbdDevicesRaw[1], "\"")
	fmt.Printf("sbd devices:%s: ", sbdDevices)

	return nil
}
