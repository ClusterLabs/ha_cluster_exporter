package main

import (
	"testing"
)

func TestDrbdParsing(t *testing.T) {
	var drbdDataRaw = []byte(` [
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
			} ],
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
				  "percent-in-sync": 100.00
				} ]
			} ]
		}
		,
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
			} ],
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
				  "percent-in-sync": 100.00
				} ]
			} ]
		}]`)

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

	if 0 != drbdDevs[0].Devices[0].Volume {
		t.Errorf("volumes should be 0")
	}
}

func TestDrbdInfoError(t *testing.T) {
	_, err := getDrbdInfo() // should fail because test environment doesn't have the drbdsetup binary
	if err == nil {
		t.Errorf("a non nil error was expected")
	}
}
