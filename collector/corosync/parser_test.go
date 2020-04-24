package corosync

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	p := NewParser()

	cfgToolOutput := []byte(`Printing ring status.
Local node ID 1084780051
RING ID 0
		id      = 10.0.0.1
		status  = ring 0 active with no faults
RING ID 1
		id      = 172.16.0.1
		status  = ring 1 active with no faults`)

	quoromToolOutput := []byte(`Quorum information
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
1084780052          1 dma-dog-hana02`)

	status, err := p.Parse(cfgToolOutput, quoromToolOutput)
	assert.NoError(t, err)

	rings := status.Rings

	assert.Len(t, rings, 2)
	assert.Equal(t, "0", rings[0].Number)
	assert.Equal(t, "10.0.0.1", rings[0].Address)
	assert.False(t, rings[0].Faulty)
	assert.Equal(t, "1", rings[1].Number)
	assert.Equal(t, "172.16.0.1", rings[1].Address)
	assert.False(t, rings[1].Faulty)

	assert.True(t, status.Quorate)
	assert.Equal(t, "1084780051", status.NodeId)
	assert.Equal(t, "1084780051", status.RingId)
	assert.Exactly(t, uint64(44), status.Seq)
	assert.Equal(t, uint64(232), status.QuorumVotes.ExpectedVotes)
	assert.Equal(t, uint64(22), status.QuorumVotes.HighestExpected)
	assert.Equal(t, uint64(21), status.QuorumVotes.TotalVotes)
	assert.Equal(t, uint64(421), status.QuorumVotes.Quorum)
}

func TestParseFaultyRings(t *testing.T) {
	cfgToolOutput := []byte(`Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults`)

	rings := parseRings(cfgToolOutput)

	assert.Len(t, rings, 2)
	assert.True(t, rings[0].Faulty)
	assert.False(t, rings[1].Faulty)
}

func TestParseNodeIdEmptyError(t *testing.T) {
	quoromToolOutput := []byte(``)

	_, err := parseNodeId(quoromToolOutput)
	assert.EqualError(t, err, "could not find Node ID line")
}

func TestParseNoQuorate(t *testing.T) {
	quoromToolOutput := []byte(`Quorate: No`)

	quorate, err := parseQuorate(quoromToolOutput)
	assert.NoError(t, err)
	assert.False(t, quorate)
}

func TestParseQuorateEmptyError(t *testing.T) {
	quoromToolOutput := []byte(``)

	_, err := parseQuorate(quoromToolOutput)
	assert.EqualError(t, err, "could not find Quorate line")
}

func TestParseQuorumVotesEmptyError(t *testing.T) {
	quoromToolOutput := []byte(``)

	_, err := parseQuoromVotes(quoromToolOutput)
	assert.EqualError(t, err, "could not find quorum votes numbers")
}

func TestParseRingIdEmptyError(t *testing.T) {
	quoromToolOutput := []byte(``)

	_, _, err := parseRingIdAndSeq(quoromToolOutput)
	assert.EqualError(t, err, "could not find Ring ID line")
}

func TestParseSeqUintError(t *testing.T) {
	quoromToolOutput := []byte(`Ring ID:          1084780051/10000000000000000000000000000000000000000000000`)

	_, _, err := parseRingIdAndSeq(quoromToolOutput)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse seq number to uint64")
	assert.Contains(t, err.Error(), "value out of range")
}

func TestParseQuorumVotesUintErrors(t *testing.T) {
	quorumToolOutputs := [][]byte{
		[]byte(`
Expected votes:   10000000000000000000000000000000000000000000000
Highest expected: 1
Total votes:      1
Quorum:           1
`),
		[]byte(`
Expected votes:   1
Highest expected: 10000000000000000000000000000000000000000000000
Total votes:      1
Quorum:           1
`),
		[]byte(`
Expected votes:   1
Highest expected: 1
Total votes:      10000000000000000000000000000000000000000000000
Quorum:           1
`),
		[]byte(`
Expected votes:   1
Highest expected: 1
Total votes:      1
Quorum:           10000000000000000000000000000000000000000000000
`),
	}
	for i, quorumToolOutput := range quorumToolOutputs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := parseQuoromVotes(quorumToolOutput)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "could not parse vote number to uint64")
			assert.Contains(t, err.Error(), "value out of range")
		})
	}
}
