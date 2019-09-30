package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Quorum metrics

// return the output of quorum in raw format
func getQuoromClusterInfo() []byte {
	// get ringStatus
	log.Println("[INFO]: Reading quorum status with corosync-quorumtool...")
	// ignore error because If any interfaces are faulty, 1 is returned by the binary. If all interfaces
	// are active 0 is returned to the shell.
	quorumInfoRaw, _ := exec.Command("/usr/sbin/corosync-quorumtool").Output()
	return quorumInfoRaw
}

func parseQuoromStatus(quoromStatus []byte) (map[string]int, string) {
	quoromRaw := string(quoromStatus)
	// Quorate:          Yes

	// Votequorum information
	// ----------------------
	// Expected votes:   2
	// Highest expected: 2
	// Total votes:      2
	// Quorum:           1

	// We apply the same method for all the metrics/data:
	// first split the string for finding the word , e.g "Expected votes:", and get it via regex
	// only the number   2,
	// and convert it to integer type
	numberOnly := regexp.MustCompile("[0-9]+")
	wordOnly := regexp.MustCompile("[a-zA-Z]+")

	quorateRaw := wordOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorate:", 2)[1])
	quorate := strings.ToLower(quorateRaw)
	expVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Expected votes:", 2)[1]))
	highVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Highest expected:", 2)[1]))
	totalVotes, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Total votes:", 2)[1]))
	quorum, _ := strconv.Atoi(numberOnly.FindString(strings.SplitAfterN(quoromRaw, "Quorum:", 2)[1]))

	voteQuorumInfo := map[string]int{
		"expectedVotes":   expVotes,
		"highestExpected": highVotes,
		"totalVotes":      totalVotes,
		"quorum":          quorum,
	}

	return voteQuorumInfo, quorate
}

// RING metrics

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
