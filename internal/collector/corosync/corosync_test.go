package corosync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"time"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestNewCorosyncCollector(t *testing.T) {
	_, err := NewCollector("../../../test/fake_corosync-cfgtool.sh", "../../../test/fake_corosync-quorumtool.sh", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	assert.Nil(t, err)
}

func TestNewCorosyncCollectorChecksCfgtoolExistence(t *testing.T) {
	_, err := NewCollector("../../../test/nonexistent", "../../../test/fake_corosync-quorumtool.sh", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestNewCorosyncCollectorChecksQuorumtoolExistence(t *testing.T) {
	_, err := NewCollector("../../../test/fake_corosync-cfgtool.sh", "../../../test/nonexistent", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestNewCorosyncCollectorChecksCfgtoolExecutableBits(t *testing.T) {
	_, err := NewCollector("../../../test/dummy", "../../../test/fake_corosync-quorumtool.sh", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestNewCorosyncCollectorChecksQuorumtoolExecutableBits(t *testing.T) {
	_, err := NewCollector("../../../test/fake_corosync-cfgtool.sh", "../../../test/dummy", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestCorosyncCollector(t *testing.T) {
	collector, _ := NewCollector("../../../test/fake_corosync-cfgtool.sh", "../../../test/fake_corosync-quorumtool.sh", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	assertcustom.Metrics(t, collector, "../../../test/corosync.metrics")
}
