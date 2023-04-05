package client

import (
	"fmt"
	"testing"
)

//
// Check error and if present, fail the test case.
// Examples usage: client.Should(t, task.Create(&r))
func Should(t *testing.T, err error) {
	if err != nil {
		t.Errorf(err.Error())
	}
}

//
// Check error and if present, fail and stop the test suite.
// Examples usage: client.Must(t, task.Create(&r))
func Must(t *testing.T, err error) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

//
// Simple equality check working for flat types (no nested types passed by reference).
func FlatEqual(got, expected interface{}) bool {
	return fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expected)
}
