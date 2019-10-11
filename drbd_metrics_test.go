package main

import (
	"fmt"
	"testing"
)

// this test verify that we return something when call the function.
// we can't really test it since we need the binary
func TestDrbdStatusFuncMinimalError(t *testing.T) {
	fmt.Println("=== Testing DRBD : testing function to get infos")
	getDrbdInfo()
}
