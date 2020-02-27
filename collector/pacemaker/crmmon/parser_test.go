package crmmon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	p := NewCrmMonParser("foo")
	assert.Equal(t, "foo", p.crmMonPath)
}

func TestParse(t *testing.T) {
	p := NewCrmMonParser("../../../test/fake_crm_mon.sh")
	data, err := p.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", data.Version)
	assert.Equal(t, 8, data.Summary.Resources.Number)
	assert.Equal(t, 1, data.Summary.Resources.Disabled)
	assert.Equal(t, 0, data.Summary.Resources.Blocked)
	assert.Equal(t, "Fri Oct 18 11:48:22 2019", data.Summary.LastChange.Time)
	assert.Equal(t, 2, data.Summary.Nodes.Number)
	assert.Equal(t, "hana01", data.Nodes[0].Name)
	assert.Equal(t, "1084783375", data.Nodes[0].Id)
	assert.Equal(t, true, data.Nodes[0].Online)
	assert.Equal(t, true, data.Nodes[0].ExpectedUp)
	assert.Equal(t, true, data.Nodes[0].DC)
	assert.Equal(t, false, data.Nodes[0].Unclean)
	assert.Equal(t, false, data.Nodes[0].Shutdown)
	assert.Equal(t, false, data.Nodes[0].StandbyOnFail)
	assert.Equal(t, false, data.Nodes[0].Maintenance)
	assert.Equal(t, false, data.Nodes[0].Pending)
	assert.Equal(t, false, data.Nodes[0].Standby)
	assert.Equal(t, "hana02", data.Nodes[1].Name)
	assert.Equal(t, "1084783376", data.Nodes[1].Id)
	assert.Equal(t, true, data.Nodes[1].Online)
	assert.Equal(t, true, data.Nodes[1].ExpectedUp)
	assert.Equal(t, false, data.Nodes[1].DC)
	assert.Equal(t, false, data.Nodes[1].Unclean)
	assert.Equal(t, false, data.Nodes[1].Shutdown)
	assert.Equal(t, false, data.Nodes[1].StandbyOnFail)
	assert.Equal(t, false, data.Nodes[1].Maintenance)
	assert.Equal(t, false, data.Nodes[1].Pending)
	assert.Equal(t, false, data.Nodes[1].Standby)
	assert.Equal(t, "hana01", data.NodeHistory.Node[0].Name)
	assert.Equal(t, 5000, data.NodeHistory.Node[0].ResourceHistory[0].MigrationThreshold)
	assert.Equal(t, 2, data.NodeHistory.Node[0].ResourceHistory[1].FailCount)
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00", data.NodeHistory.Node[0].ResourceHistory[0].Name)
	assert.Equal(t, 1, len(data.Resources))
	assert.Equal(t, "test-stop", data.Resources[0].Id)
	assert.Equal(t, false, data.Resources[0].Active)
	assert.Equal(t, "Stopped", data.Resources[0].Role)
}
