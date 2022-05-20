package pacemaker

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestNewPacemakerCollector(t *testing.T) {
	_, err := NewCollector("../../test/fake_crm_mon.sh", "../../test/fake_cibadmin.sh", false, log.NewNopLogger())

	assert.Nil(t, err)
}

func TestNewPacemakerCollectorChecksCrmMonExistence(t *testing.T) {
	_, err := NewCollector("../../test/nonexistent", "", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/nonexistent' does not exist")
}

func TestNewPacemakerCollectorChecksCrmMonExecutableBits(t *testing.T) {
	_, err := NewCollector("../../test/dummy", "", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/dummy' is not executable")
}

func TestPacemakerCollector(t *testing.T) {
	collector, err := NewCollector("../../test/fake_crm_mon.sh", "../../test/fake_cibadmin.sh", false, log.NewNopLogger())

	assert.Nil(t, err)
	assertcustom.Metrics(t, collector, "pacemaker.metrics")
}
