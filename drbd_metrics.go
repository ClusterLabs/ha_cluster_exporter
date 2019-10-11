package main

import (
	"log"
	"os/exec"
)

// drbdStatus is for parsing relevant data we want to convert to metrics
type drbdStatus struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Devices []struct {
		Volume int `json:"volume"`
	} `json:"devices"`
}

// return drbd status in byte raw json
func getDrbdInfo() []byte {
	// get ringStatus
	log.Println("[INFO]: Reading drbd status with drbdsetup status ...")
	drbdStatusRaw, _ := exec.Command("/sbin/drbdsetup", "status", "--json").Output()
	return drbdStatusRaw
}
