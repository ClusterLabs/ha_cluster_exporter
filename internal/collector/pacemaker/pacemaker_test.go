package pacemaker

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestNewPacemakerCollector(t *testing.T) {
	_, err := NewCollector("../../../test/fake_crm_mon.sh", "../../../test/fake_cibadmin.sh", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.Nil(t, err)
}

func TestNewPacemakerCollectorChecksCrmMonExistence(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewCollector("../../../test/nonexistent", "", 10*time.Second, logger)

	assert.NoError(t, err)
}

func TestNewPacemakerCollectorChecksCrmMonExecutableBits(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewCollector("../../../test/dummy", "", 10*time.Second, logger)

	assert.NoError(t, err)
}

func TestPacemakerCollector(t *testing.T) {
	// Force Local time to UTC for deterministic test results
	origLocal := time.Local
	time.Local = time.UTC
	defer func() { time.Local = origLocal }()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	collector, err := NewCollector("../../../test/fake_crm_mon.sh", "../../../test/fake_cibadmin.sh", 10*time.Second, logger)

	assert.Nil(t, err)
	assertcustom.Metrics(t, collector, "../../../test/pacemaker.metrics")
}
