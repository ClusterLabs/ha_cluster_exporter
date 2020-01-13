package main

import (
	"testing"
)

// TEST group quorum metrics
func TestQuoromMetricParsing(t *testing.T) {
	// the data is fake
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
	`
	voteQuorumInfo, quorate, _ := parseQuoromStatus([]byte(quoromStatus))

	if voteQuorumInfo["expected_votes"] != 232 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["expectedVotes"])
	}
	if voteQuorumInfo["highest_expected"] != 22 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["highestExpected"])
	}

	if voteQuorumInfo["total_votes"] != 21 {
		t.Errorf("expectedVotes should be 232 got instead: %d", voteQuorumInfo["totalVotes"])
	}

	if voteQuorumInfo["quorum"] != 421 {
		t.Errorf("expectedVotes should be 421 got instead: %d", voteQuorumInfo["quorum"])
	}

	if quorate != 1 {
		t.Errorf("quorate should be 1, got %v", quorate)
	}
}

// TEST group RING metrics
// test that we recognize 1 error (for increasing metric later)
func TestOneRingError(t *testing.T) {
	ringStatusWithOneError := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults																				   
			`

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
	ringStatusWithOneError := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults																				   
			`

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
	ringErrorsTotal, err := parseRingStatus([]byte(ringStatusWithOneError))
	if err != nil {
		t.Error(err)
	}

	RingExpectedErrors := 3
	if ringErrorsTotal != RingExpectedErrors {
		t.Errorf("ringErrors was incorrect, got: %d, expected: %d.", ringErrorsTotal, RingExpectedErrors)
	}
}

func TestRingStatusParsingError(t *testing.T) {
	_, err := parseRingStatus([]byte("some error occurred"))
	if err == nil {
		t.Fatal("a non nil error was expected")
	}
	if err.Error() != "corosync-cfgtool returned unexpected output: some error occurred" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewCorosyncCollector(t *testing.T) {
	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/fake_corosync-quorumtool.sh")
	if err != nil {
		t.Errorf("Unexpected error, got: %v", err)
	}
}

func TestNewCorosyncCollectorChecksCfgtoolExistence(t *testing.T) {
	_, err := NewCorosyncCollector("test/nonexistent", "test/fake_corosync-quorumtool.sh")
	if err == nil {
		t.Fatal("a non nil error was expected")
	}
	if err.Error() != "could not initialize Corosync collector: 'test/nonexistent' does not exist" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewCorosyncCollectorChecksQuorumtoolExistence(t *testing.T) {

	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/nonexistent")
	if err == nil {
		t.Fatal("a non nil error was expected")
	}
	if err.Error() != "could not initialize Corosync collector: 'test/nonexistent' does not exist" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewCorosyncCollectorChecksCfgtoolExecutableBits(t *testing.T) {
	_, err := NewCorosyncCollector("test/dummy", "test/fake_corosync-quorumtool.sh")
	if err == nil {
		t.Fatal("a non nil error was expected")
	}
	if err.Error() != "could not initialize Corosync collector: 'test/dummy' is not executable" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewCorosyncCollectorChecksQuorumtoolExecutableBits(t *testing.T) {
	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/dummy")
	if err == nil {
		t.Fatal("a non nil error was expected")
	}
	if err.Error() != "could not initialize Corosync collector: 'test/dummy' is not executable" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCorosyncCollector(t *testing.T) {
	collector, _ := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/fake_corosync-quorumtool.sh")
	expectMetrics(t, collector, "corosync.metrics")
}
