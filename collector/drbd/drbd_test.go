package drbd

import (
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestDrbdParsing(t *testing.T) {
	var drbdDataRaw = []byte(`
[
  {
    "name": "1-single-0",
    "node-id": 2,
    "role": "Secondary",
    "suspended": false,
    "write-ordering": "flush",
    "devices": [
      {
        "volume": 0,
        "minor": 2,
        "disk-state": "UpToDate",
        "client": false,
        "quorum": true,
        "size": 409600,
        "read": 654321,
        "written": 123456,
        "al-writes": 123,
        "bm-writes": 321,
        "upper-pending": 1,
        "lower-pending": 2
      }
    ],
    "connections": [
      {
        "peer-node-id": 1,
        "name": "SLE15-sp1-gm-drbd1145296-node1",
        "connection-state": "Connected",
        "congested": false,
        "peer-role": "Primary",
        "ap-in-flight": 0,
        "rs-in-flight": 0,
        "peer_devices": [
          {
            "volume": 0,
            "replication-state": "Established",
            "peer-disk-state": "UpToDate",
            "peer-client": false,
            "resync-suspended": "no",
            "received": 456,
            "sent": 654,
            "out-of-sync": 0,
            "pending": 3,
            "unacked": 4,
            "has-sync-details": false,
            "has-online-verify-details": false,
            "percent-in-sync": 100
          }
        ]
      }
    ]
  },
  {
    "name": "1-single-1",
    "node-id": 2,
    "role": "Secondary",
    "suspended": false,
    "write-ordering": "flush",
    "devices": [
      {
        "volume": 0,
        "minor": 3,
        "disk-state": "UpToDate",
        "client": false,
        "quorum": false,
        "size": 10200,
        "read": 654321,
        "written": 123456,
        "al-writes": 123,
        "bm-writes": 321,
        "upper-pending": 1,
        "lower-pending": 2
      }
    ],
    "connections": [
      {
        "peer-node-id": 1,
        "name": "SLE15-sp1-gm-drbd1145296-node1",
        "connection-state": "Connected",
        "congested": false,
        "peer-role": "Primary",
        "ap-in-flight": 0,
        "rs-in-flight": 0,
        "peer_devices": [
          {
            "volume": 0,
            "replication-state": "Established",
            "peer-disk-state": "UpToDate",
            "peer-client": false,
            "resync-suspended": "no",
            "received": 456,
            "sent": 654,
            "out-of-sync": 0,
            "pending": 3,
            "unacked": 4,
            "has-sync-details": false,
            "has-online-verify-details": false,
            "percent-in-sync": 99.8
          }
        ]
      }
    ]
  }
]`)

	drbdDevs, err := parseDrbdStatus(drbdDataRaw)

	assert.Nil(t, err)
	assert.Equal(t, "1-single-0", drbdDevs[0].Name)
	assert.Equal(t, "Secondary", drbdDevs[0].Role)
	assert.Equal(t, "UpToDate", drbdDevs[0].Devices[0].DiskState)
	assert.Equal(t, 1, drbdDevs[0].Connections[0].PeerNodeID)
	assert.Equal(t, "UpToDate", drbdDevs[0].Connections[0].PeerDevices[0].PeerDiskState)
	assert.Equal(t, 0, drbdDevs[0].Devices[0].Volume)
	assert.Equal(t, 123456, drbdDevs[0].Devices[0].Written)
	assert.Equal(t, 654321, drbdDevs[0].Devices[0].Read)
	assert.Equal(t, 123, drbdDevs[0].Devices[0].AlWrites)
	assert.Equal(t, 321, drbdDevs[0].Devices[0].BmWrites)
	assert.Equal(t, 1, drbdDevs[0].Devices[0].UpPending)
	assert.Equal(t, 2, drbdDevs[0].Devices[0].LoPending)
	assert.Equal(t, true, drbdDevs[0].Devices[0].Quorum)
	assert.Equal(t, false, drbdDevs[1].Devices[0].Quorum)
	assert.Equal(t, 456, drbdDevs[0].Connections[0].PeerDevices[0].Received)
	assert.Equal(t, 654, drbdDevs[0].Connections[0].PeerDevices[0].Sent)
	assert.Equal(t, 3, drbdDevs[0].Connections[0].PeerDevices[0].Pending)
	assert.Equal(t, 4, drbdDevs[0].Connections[0].PeerDevices[0].Unacked)
	assert.Equal(t, 100.0, drbdDevs[0].Connections[0].PeerDevices[0].PercentInSync)
	assert.Equal(t, 99.8, drbdDevs[1].Connections[0].PeerDevices[0].PercentInSync)
}

func TestNewDrbdCollector(t *testing.T) {
	_, err := NewCollector("../../test/fake_drbdsetup.sh", "splitbrainpath", false, log.NewNopLogger())

	assert.Nil(t, err)
}

func TestNewDrbdCollectorChecksDrbdsetupExistence(t *testing.T) {
	_, err := NewCollector("../../test/nonexistent", "splitbrainfake", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/nonexistent' does not exist")
}

func TestNewDrbdCollectorChecksDrbdsetupExecutableBits(t *testing.T) {
	_, err := NewCollector("../../test/dummy", "splibrainfake", false, log.NewNopLogger())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'../../test/dummy' is not executable")
}

func TestDRBDCollector(t *testing.T) {
	collector, _ := NewCollector("../../test/fake_drbdsetup.sh", "fake", false, log.NewNopLogger())
	assertcustom.Metrics(t, collector, "drbd.metrics")
}

func TestDRBDSplitbrainCollector(t *testing.T) {
	collector, _ := NewCollector("../../test/fake_drbdsetup.sh", "../../test/drbd-splitbrain", false, log.NewNopLogger())

	expect := `
	# HELP ha_cluster_drbd_split_brain Whether a split brain has been detected; 1 line per resource, per volume.
	# TYPE ha_cluster_drbd_split_brain gauge
	ha_cluster_drbd_split_brain{resource="resource01",volume="vol01"} 1
	ha_cluster_drbd_split_brain{resource="resource02",volume="vol02"} 1
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expect), "ha_cluster_drbd_split_brain")

	assert.NoError(t, err)
}
