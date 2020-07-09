package pacemaker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
)

func TestNewPacemakerCollector(t *testing.T) {
	_, err := NewCollector("../../test/fake_crm_mon.sh", "../../test/fake_cibadmin.sh")

	assert.Nil(t, err)
}

func TestNewPacemakerCollectorChecksCrmMonExistence(t *testing.T) {
	_, err := NewCollector("../../test/nonexistent", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/nonexistent' does not exist")
}

func TestNewPacemakerCollectorChecksCrmMonExecutableBits(t *testing.T) {
	_, err := NewCollector("../../test/dummy", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/dummy' is not executable")
}

func TestPacemakerCollector(t *testing.T) {
	collector, err := NewCollector("../../test/fake_crm_mon.sh", "../../test/fake_cibadmin.sh")
	assert.NoError(t, err)
	collector.Clock = &clock.StoppedClock{}

	assertcustom.Metrics(t, collector, "pacemaker.metrics")
}
