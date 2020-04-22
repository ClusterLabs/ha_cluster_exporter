package cfgtool

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
			status  = Marking ringid 0 interface 10.0.0.1
	RING ID 1
			id      = 172.16.0.1
			status  = ring 1 active with no faults`

	status, err := p.Parse([]byte(input))
	assert.NoError(t, err)

	assert.Equal(t, "16777226", status.NodeId)

	rings := status.Rings

	assert.Len(t, rings, 2)
	assert.Equal(t, "0", rings[0].Id)
	assert.Equal(t, "10.0.0.1", rings[0].Address)
	assert.False(t, rings[0].Faulty)
	assert.Equal(t, "1", rings[1].Id)
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

// test that we recognize 3 rings error (for increasing metric later)
func TestMultipleRingErrors(t *testing.T) {
	p := NewParser()

	input := `Printing ring status.
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

	status, err := p.Parse([]byte(input))
	assert.NoError(t, err)

	rings := status.Rings

	assert.Len(t, rings, 6)
}
