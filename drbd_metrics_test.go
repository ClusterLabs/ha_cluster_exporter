package main

import (
	"testing"
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
        "read": 0,
        "written": 548525,
        "al-writes": 4,
        "bm-writes": 0,
        "upper-pending": 0,
        "lower-pending": 0
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
            "received": 548525,
            "sent": 0,
            "out-of-sync": 0,
            "pending": 0,
            "unacked": 0,
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
        "quorum": true,
        "size": 10200,
        "read": 0,
        "written": 546005,
        "al-writes": 1,
        "bm-writes": 0,
        "upper-pending": 0,
        "lower-pending": 0
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
            "received": 546005,
            "sent": 0,
            "out-of-sync": 0,
            "pending": 0,
            "unacked": 0,
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
	if err != nil {
		t.Error(err)
	}

	// test attributes
	if "1-single-0" != drbdDevs[0].Name {
		t.Errorf("name doesn't correspond! fail got %s", drbdDevs[0].Name)
	}

	if "Secondary" != drbdDevs[0].Role {
		t.Errorf("role doesn't correspond! fail got %s", drbdDevs[0].Role)
	}

	if "UpToDate" != drbdDevs[0].Devices[0].DiskState {
		t.Errorf("disk-states doesn't correspond! fail got %s", drbdDevs[0].Devices[0].DiskState)
	}

	if 1 != drbdDevs[0].Connections[0].PeerNodeID {
		t.Errorf("peerNodeID doesn't correspond! fail got %d", drbdDevs[0].Connections[0].PeerNodeID)
	}

	if "UpToDate" != drbdDevs[0].Connections[0].PeerDevices[0].PeerDiskState {
		t.Errorf("peerDiskState doesn't correspond! fail got %s", drbdDevs[0].Connections[0].PeerDevices[0].PeerDiskState)
	}

	if 0 != drbdDevs[0].Devices[0].Volume {
		t.Errorf("volumes should be 0")
	}

	if 100 != drbdDevs[0].Connections[0].PeerDevices[0].PercentInSync {
		t.Errorf("PercentInSync doesn't correspond! fail got %f", drbdDevs[0].Connections[0].PeerDevices[0].PercentInSync)
	}

	if 99.8 != drbdDevs[1].Connections[0].PeerDevices[0].PercentInSync {
		t.Errorf("Float PercentInSync doesn't correspond! fail got %f", drbdDevs[1].Connections[0].PeerDevices[0].PercentInSync)
	}

}

func TestDrbdInfoError(t *testing.T) {
	drbdsetupPath = "test/nonexistent"
	_, err := getDrbdInfo()
	if err == nil {
		t.Errorf("a non nil error was expected")
	}
}

func TestNewDrbdCollector(t *testing.T) {
	drbdsetupPath = "test/fake_drbdsetup.sh"

	_, err := NewDrbdCollector()
	if err != nil {
		t.Errorf("Unexpected error, got: %v", err)
	}
}

func TestNewDrbdCollectorChecksDrbdsetupExistence(t *testing.T) {
	drbdsetupPath = "test/nonexistent"

	_, err := NewDrbdCollector()
	if err == nil {
		t.Error("a non nil error was expected")
	}
	if err.Error() != "'test/nonexistent' not found: stat test/nonexistent: no such file or directory" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNewDrbdCollectorChecksDrbdsetupExecutableBits(t *testing.T) {
	drbdsetupPath = "test/dummy"

	_, err := NewDrbdCollector()
	if err == nil {
		t.Fatalf("a non nil error was expected")
	}
	if err.Error() != "'test/dummy' is not executable" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDRBDCollector(t *testing.T) {
	clock = StoppedClock{}
	drbdsetupPath = "test/fake_drbdsetup.sh"

	collector, _ := NewDrbdCollector()
	expectMetrics(t, collector, "drbd.metrics")
}
