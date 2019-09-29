package main

import (
	"fmt"
	"testing"
)

// TEST group quorum metrics
func TestQuoromMetricParsing(t *testing.T) {
	// the data is fake data
	fmt.Println("=== Testing quorum info retrivial")
	quoromStatus := `
	Quorum information
	------------------
	Date:             Sun Sep 29 16:10:37 2019
	Quorum provider:  corosync_votequorum
	Nodes:            2
	Node ID:          1084780051
	Ring ID:          1084780051/44
	Quorate:          Yes
	
	Votequorum information
	----------------------
	Expected votes:   232
	Highest expected: 22
	Total votes:      21
	Quorum:           421  
	Flags:            2Node Quorate WaitForAll 
	
	Membership information
	----------------------
		Nodeid      Votes Name
	1084780051          1 dma-dog-hana01 (local)
	1084780052          1 dma-dog-hana02
	dma-dog-hana01:~ # 
	`
	getQuoromClusterInfo()
	voteQuorumInfo, quorate := parseQuoromStatus([]byte(quoromStatus))

	if voteQuorumInfo["expectedVotes"] != 232 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["expectedVotes"])
	}
	if voteQuorumInfo["highestExpected"] != 22 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["highestExpected"])
	}

	if voteQuorumInfo["totalVotes"] != 21 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["totalVotes"])
	}

	if voteQuorumInfo["quorum"] != 421 {
		t.Errorf("expectedVotes should be 421 got instead: %d", voteQuorumInfo["quorum"])
	}

	if quorate != "Yes" {
		t.Errorf("quorate should be set to Yes, got %s", quorate)
	}

}

// TEST group RING metrics
// test that we recognize 1 error (for increasing metric later)
func TestOneRingError(t *testing.T) {
	fmt.Println("=== Test one ring error")
	ringStatusWithOneError := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults																				   
			`

	getCorosyncRingStatus()
	ringErrorsTotal, err := parseRingStatus([]byte(ringStatusWithOneError))
	RingExpectedErrors := 1
	if ringErrorsTotal != RingExpectedErrors {
		t.Errorf("ringErrors was incorrect, got: %d, expected: %d.", ringErrorsTotal, RingExpectedErrors)
	}
	if err != nil {
		t.Errorf("error should be nil got instead: %s", err)
	}
}

func TestZeroRingErrors(t *testing.T) {
	fmt.Println("=== Test zero Ring errors")
	ringStatusWithOneError := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults																				   
			`

	getCorosyncRingStatus()
	ringErrorsTotal, err := parseRingStatus([]byte(ringStatusWithOneError))
	RingExpectedErrors := 0
	if ringErrorsTotal != RingExpectedErrors {
		t.Errorf("ringErrors was incorrect, got: %d, expected: %d.", ringErrorsTotal, RingExpectedErrors)
	}
	if err != nil {
		t.Errorf("error should be nil got instead: %s", err)
	}
}

// test that we recognize 3 rings error (for increasing metric later)
func TestMultipleRingErrors(t *testing.T) {
	fmt.Println("=== Test multiples ring error")
	ringStatusWithOneError := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults
	RING ID 2
			id      = 10.0.0.1
			status  = Marking ringid 1 interface 10.0.0.1 FAULTY
	RING ID 3
			id      = 172.16.0.1
			status  = ring 1 active with no faults
	RING ID 4
			id      = 10.0.0.1
			status  = Marking ringid 1 interface 10.0.0.1 FAULTY
	RING ID 5
			id      = 172.16.0.1
			status  = ring 1 active with no faults
																											   
			`

	getCorosyncRingStatus()
	ringErrorsTotal, err := parseRingStatus([]byte(ringStatusWithOneError))
	RingExpectedErrors := 3
	if ringErrorsTotal != RingExpectedErrors {
		t.Errorf("ringErrors was incorrect, got: %d, expected: %d.", ringErrorsTotal, RingExpectedErrors)
	}
	if err != nil {
		t.Errorf("error should be nil got instead: %s", err)
	}
}

// test that in case of system unexpected error we detect this
func TestSystemUnexpectedError(t *testing.T) {
	fmt.Println("=== Test unexpected error")
	// since there is no cluster in a Test env. this will return an error
	ringStatusError := getCorosyncRingStatus()
	parseRingStatus([]byte(ringStatusError))
	ringErrorsTotal, err := parseRingStatus([]byte(ringStatusError))
	RingExpectedErrors := 0
	if ringErrorsTotal != RingExpectedErrors {
		t.Errorf("ringErrors was incorrect, got: %d, expected: %d.", ringErrorsTotal, RingExpectedErrors)
	}
	if err == nil {
		t.Errorf("error should not be nil got !!")
	}

}
