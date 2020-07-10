package collector

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/ClusterLabs/ha_cluster_exporter/internal/clock"
	"github.com/ClusterLabs/ha_cluster_exporter/test/mock_collector"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestInstrumentedCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollector := mock_collector.NewMockInstrumentableCollector(ctrl)
	mockCollector.EXPECT().GetSubsystem().Return("mock_collector").AnyTimes()
	mockCollector.EXPECT().Describe(gomock.Any())
	mockCollector.EXPECT().CollectWithError(gomock.Any())

	SUT := NewInstrumentedCollector(mockCollector)
	SUT.Clock = &clock.StoppedClock{}

	metrics := `# HELP ha_cluster_scrape_duration_seconds Duration of a collector scrape.
# TYPE ha_cluster_scrape_duration_seconds gauge
ha_cluster_scrape_duration_seconds{collector="mock_collector"} 1.234
# HELP ha_cluster_scrape_success Whether a collector succeeded.
# TYPE ha_cluster_scrape_success gauge
ha_cluster_scrape_success{collector="mock_collector"} 1
`

	err := testutil.CollectAndCompare(SUT, strings.NewReader(metrics))
	assert.NoError(t, err)
}

func TestInstrumentedCollectorScrapeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logHook := testlog.NewGlobal()
	defer logHook.Reset()

	mockCollector := mock_collector.NewMockInstrumentableCollector(ctrl)
	mockCollector.EXPECT().GetSubsystem().Return("mock_collector").AnyTimes()
	mockCollector.EXPECT().Describe(gomock.Any())
	mockCollector.EXPECT().CollectWithError(gomock.Any()).Return(errors.New("test error"))

	SUT := NewInstrumentedCollector(mockCollector)

	metrics := `# HELP ha_cluster_scrape_success Whether a collector succeeded.
# TYPE ha_cluster_scrape_success gauge
ha_cluster_scrape_success{collector="mock_collector"} 0
`

	err := testutil.CollectAndCompare(SUT, strings.NewReader(metrics), "ha_cluster_scrape_success")
	assert.NoError(t, err)

	assert.Len(t, logHook.Entries, 1)
	assert.Contains(t, logHook.LastEntry().Message, "test error")
}
