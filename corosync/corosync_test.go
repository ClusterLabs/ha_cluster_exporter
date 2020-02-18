package corosync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
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

	assert.Equal(t, 232, voteQuorumInfo["expected_votes"])
	assert.Equal(t, 22, voteQuorumInfo["highest_expected"])
	assert.Equal(t, 21, voteQuorumInfo["total_votes"])
	assert.Equal(t, 421, voteQuorumInfo["quorum"])
	assert.Equal(t, 1.0, quorate)
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

	assert.Nil(t, err)
	assert.Equal(t, 1, ringErrorsTotal)
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

	assert.Nil(t, err)
	assert.Equal(t, 0, ringErrorsTotal)
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

	assert.Nil(t, err)
	assert.Equal(t, 3, ringErrorsTotal)
}

func TestRingStatusParsingError(t *testing.T) {
	_, err := parseRingStatus([]byte("some error occurred"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some error occurred")
}

func TestNewCorosyncCollector(t *testing.T) {
	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/fake_corosync-quorumtool.sh")
	assert.Nil(t, err)
}

func TestNewCorosyncCollectorChecksCfgtoolExistence(t *testing.T) {
	_, err := NewCorosyncCollector("test/nonexistent", "test/fake_corosync-quorumtool.sh")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'test/nonexistent' does not exist")
}

func TestNewCorosyncCollectorChecksQuorumtoolExistence(t *testing.T) {
	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'test/nonexistent' does not exist")
}

func TestNewCorosyncCollectorChecksCfgtoolExecutableBits(t *testing.T) {
	_, err := NewCorosyncCollector("test/dummy", "test/fake_corosync-quorumtool.sh")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'test/dummy' is not executable")
}

func TestNewCorosyncCollectorChecksQuorumtoolExecutableBits(t *testing.T) {
	_, err := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/dummy")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'test/dummy' is not executable")
}

func TestCorosyncCollector(t *testing.T) {
	collector, _ := NewCorosyncCollector("test/fake_corosync-cfgtool.sh", "test/fake_corosync-quorumtool.sh")
	assertcustom.Metrics(t, collector, "corosync.metrics")
}
