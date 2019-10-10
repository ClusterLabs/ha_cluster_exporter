package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// return drbd status in byte raw json
func getDrbdInfo() []byte {
	// get ringStatus
	log.Println("[INFO]: Reading drbd status with drbdsetup status ...")
	drbdStatusRaw, _ := exec.Command("/usr/sbin/drbdsetup", "status" "--json").Output()
	return drbdStatusRaw
}
