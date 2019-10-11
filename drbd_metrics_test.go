package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

// this test verify that we return something when call the function.
// we can't really test it since we need the binary
func TestDrbdStatusFuncMinimalError(t *testing.T) {
	fmt.Println("=== Testing DRBD : testing function to get infos")
	getDrbdInfo()
}

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

	var drbdDevs []drbdStatus
	err := json.Unmarshal(drbdDataRaw, &drbdDevs)
	if err != nil {
		log.Fatalln("[ERROR]:", err)
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
