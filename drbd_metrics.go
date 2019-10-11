package main

import (
	"log"
	"os/exec"
)

// return drbd status in byte raw json
func getDrbdInfo() []byte {
	// get ringStatus
	log.Println("[INFO]: Reading drbd status with drbdsetup status ...")
	drbdStatusRaw, _ := exec.Command("/sbin/drbdsetup", "status", "--json").Output()
	return drbdStatusRaw
}
