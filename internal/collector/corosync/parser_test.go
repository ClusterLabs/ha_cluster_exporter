package corosync

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	p := NewParser()

	cfgToolOutput := []byte(`Printing link status.
Local node ID 1084780051
Link ID 0
		addr	= 10.0.0.1
		status	= OK
Link ID 1
        addr      = 172.16.0.1
        status    = OK`)

	quoromToolOutput := []byte(`Quorum information
------------------
Date:             Sun Sep 29 16:10:37 2019
Quorum provider:  corosync_votequorum
Nodes:            2
Node ID:          1084780051
Ring ID:          1084780051.44
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
	Nodeid      Votes Qdevice Name
1084780051          1      NR dma-dog-hana01 (local)
1084780052          1      A,V,NMW dma-dog-hana02`)

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
	assert.Equal(t, "1084780051.44", status.RingId)
	assert.EqualValues(t, 232, status.QuorumVotes.ExpectedVotes)
	assert.EqualValues(t, 22, status.QuorumVotes.HighestExpected)
	assert.EqualValues(t, 21, status.QuorumVotes.TotalVotes)
	assert.EqualValues(t, 421, status.QuorumVotes.Quorum)

	members := status.Members
	assert.Len(t, members, 2)
	assert.Exactly(t, "1084780051", members[0].Id)
	assert.Exactly(t, "dma-dog-hana01", members[0].Name)
	assert.Exactly(t, "NR", members[0].Qdevice)
	assert.True(t, members[0].Local)
	assert.EqualValues(t, 1, members[0].Votes)
	assert.Exactly(t, "1084780052", members[1].Id)
	assert.Exactly(t, "dma-dog-hana02", members[1].Name)
	assert.Exactly(t, "A,V,NMW", members[1].Qdevice)
	assert.False(t, members[1].Local)
	assert.EqualValues(t, 1, members[1].Votes)
}

func TestParseRingIdInCorosyncV2_4(t *testing.T) {
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

	ringId, err := parseRingId(quoromToolOutput)
	assert.NoError(t, err)

	assert.Equal(t, "1084780051/44", ringId)
}

func TestParseRingIdInCorosyncV2_3(t *testing.T) {
	quoromToolOutput := []byte(`Quorum information
------------------
Date:             Wed May 27 14:16:10 2020
Quorum provider:  corosync_votequorum
Nodes:            2
Node ID:          1
Ring ID:          100
Quorate:          Yes
Votequorum information
----------------------
Expected votes:   2
Highest expected: 2
Total votes:      2
Quorum:           1
Flags:            2Node Quorate WaitForAll
Membership information
----------------------
    Nodeid      Votes Name
         1          1 10.1.2.4 (local)
         2          1 10.1.2.5`)

	ringId, err := parseRingId(quoromToolOutput)
	assert.NoError(t, err)

	assert.Equal(t, "100", ringId)
}

func TestParseFaultyRings(t *testing.T) {
	cfgToolOutput := []byte(`Printing ring status.
	Local node ID 16777226
	Link ID 0
			addr      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	Link ID 1
			addr      = 172.16.0.1
			status  = ring 1 active with no faults`)

	rings := parseRings(cfgToolOutput)

	assert.Len(t, rings, 2)
	assert.True(t, rings[0].Faulty)
	assert.False(t, rings[1].Faulty)
}

func TestParseFaultyRingsInCorosyncV2(t *testing.T) {
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

	_, err := parseRingId(quoromToolOutput)
	assert.EqualError(t, err, "could not find Ring ID line")
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

func TestParseMembersEmptyError(t *testing.T) {
	quoromToolOutput := []byte(``)

	_, err := parseMembers(quoromToolOutput)
	assert.EqualError(t, err, "could not find membership information")
}

func TestParseMembersUintError(t *testing.T) {
	quoromToolOutput := []byte(`Membership information
----------------------
    Nodeid      Votes Qdevice Name
1084780051 10000000000000000000000000000000000000000000000 NW dma-dog-hana01`)

	_, err := parseMembers(quoromToolOutput)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse vote number to uint64")
	assert.Contains(t, err.Error(), "value out of range")
}

func TestParseMembersWithAnotherExample(t *testing.T) {
	quorumToolOutput := []byte(`Quorum information
------------------
Date:             Mon May  4 16:50:13 2020
Quorum provider:  corosync_votequorum
Nodes:            2
Node ID:          2
Ring ID:          1/8
Quorate:          Yes

Votequorum information
----------------------
Expected votes:   2
Highest expected: 2
Total votes:      2
Quorum:           1  
Flags:            2Node Quorate WaitForAll 

Membership information
----------------------
    Nodeid      Votes Qdevice Name
         1          1      NR  192.168.127.20
         2          1      NR  192.168.127.21 (local)`)

	members, err := parseMembers(quorumToolOutput)

	assert.NoError(t, err)

	assert.Len(t, members, 2)
	assert.Exactly(t, "1", members[0].Id)
	assert.Exactly(t, "192.168.127.20", members[0].Name)
	assert.False(t, members[0].Local)
	assert.EqualValues(t, 1, members[0].Votes)
	assert.Exactly(t, "2", members[1].Id)
	assert.Exactly(t, "192.168.127.21", members[1].Name)
	assert.True(t, members[1].Local)
	assert.EqualValues(t, 1, members[1].Votes)
}

func TestParseMembersWithIpv6Hostnames(t *testing.T) {
	quorumToolOutput := []byte(`Quorum information
Membership information
----------------------
    Nodeid      Votes Qdevice Name
         1          1      NR  fe80:00:000:0000:1234:5678:ABCD:EF
         2          1      NR  FE80:0:00:000:0000::1 (local)`)

	members, err := parseMembers(quorumToolOutput)

	assert.NoError(t, err)

	assert.Len(t, members, 2)
	assert.Exactly(t, "1", members[0].Id)
	assert.Exactly(t, "fe80:00:000:0000:1234:5678:ABCD:EF", members[0].Name)
	assert.False(t, members[0].Local)
	assert.EqualValues(t, 1, members[0].Votes)
	assert.Exactly(t, "2", members[1].Id)
	assert.Exactly(t, "FE80:0:00:000:0000::1", members[1].Name)
	assert.True(t, members[1].Local)
	assert.EqualValues(t, 1, members[1].Votes)
}
