package corosync

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestNewCorosyncCollector(t *testing.T) {
	_, err := NewCollector("../../test/fake_corosync-cfgtool.sh", "../../test/fake_corosync-quorumtool.sh", false, log.NewNopLogger())
	assert.Nil(t, err)
}

func TestNewCorosyncCollectorChecksCfgtoolExistence(t *testing.T) {
	_, err := NewCollector("../../test/nonexistent", "../../test/fake_corosync-quorumtool.sh", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/nonexistent' does not exist")
}

func TestNewCorosyncCollectorChecksQuorumtoolExistence(t *testing.T) {
	_, err := NewCollector("../../test/fake_corosync-cfgtool.sh", "../../test/nonexistent", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/nonexistent' does not exist")
}

func TestNewCorosyncCollectorChecksCfgtoolExecutableBits(t *testing.T) {
	_, err := NewCollector("../../test/dummy", "../../test/fake_corosync-quorumtool.sh", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/dummy' is not executable")
}

func TestNewCorosyncCollectorChecksQuorumtoolExecutableBits(t *testing.T) {
	_, err := NewCollector("../../test/fake_corosync-cfgtool.sh", "../../test/dummy", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/dummy' is not executable")
}

func TestCorosyncCollector(t *testing.T) {
	collector, _ := NewCollector("../../test/fake_corosync-cfgtool.sh", "../../test/fake_corosync-quorumtool.sh", false, log.NewNopLogger())
	assertcustom.Metrics(t, collector, "corosync.metrics")
}
