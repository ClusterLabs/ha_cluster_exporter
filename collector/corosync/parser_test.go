package corosync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	p := NewParser()

	input := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = ring 0 active with no faults
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults`

	status, err := p.Parse([]byte(input))
	assert.NoError(t, err)

	assert.Equal(t, "16777226", status.NodeId)

	rings := status.Rings

	assert.Len(t, rings, 2)
	assert.Equal(t, "0", rings[0].Number)
	assert.Equal(t, "10.0.0.1", rings[0].Address)
	assert.False(t, rings[0].Faulty)
	assert.Equal(t, "1", rings[1].Number)
	assert.Equal(t, "172.16.0.1", rings[1].Address)
	assert.False(t, rings[1].Faulty)
}

func TestParseFaultyRings(t *testing.T) {
	p := NewParser()

	input := `Printing ring status.
	Local node ID 16777226
	RING ID 0
			id      = 10.0.0.1
			status  = Marking ringid 0 interface 10.0.0.1 FAULTY
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults`

	status, err := p.Parse([]byte(input))
	assert.NoError(t, err)

	rings := status.Rings

	assert.Len(t, rings, 2)
	assert.True(t, rings[0].Faulty)
	assert.False(t, rings[1].Faulty)
}
