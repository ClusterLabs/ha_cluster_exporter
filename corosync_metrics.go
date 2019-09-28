package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// get status ring and return it as bytes
// this function can return also just an malformed output in case of error, we don't check.
// It is the parser that will check the status
func getCorosyncRingStatus() []byte {
	// get ringStatus
	log.Println("[INFO]: Reading ring status with corosync-cfgtool...")
	// ignore error because If any interfaces are faulty, 1 is returned by the binary. If all interfaces
	// are active 0 is returned to the shell.
	ringStatusRaw, _ := exec.Command("/usr/sbin/corosync-cfgtool", "-s").Output()
	return ringStatusRaw
}

// return the number of RingError that we will use as gauge, and error if somethings unexpected happens
func parseRingStatus(ringStatus []byte) (int, error) {
	statusRaw := string(ringStatus)
	// check if there is a ring ERROR first
	ringErrorsTotal := strings.Count(statusRaw, "FAULTY")

	// in case there is no error we need to check that the output is not
	if ringErrorsTotal == 0 {
		// if there is no RING ID word, the command corosync-cfgtool went wrong/error out
		if strings.Count(statusRaw, "RING ID") == 0 {
			return 0, fmt.Errorf("[ERROR]: corosync-cfgtool command returned an unexpected error %s", statusRaw)
		}

		return 0, nil
	}

	// there is a ringError
	return ringErrorsTotal, nil
}
