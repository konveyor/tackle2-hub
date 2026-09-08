package assert

import (
	"testing"
)

// Check error and if present, fail the test case.
// Examples usage: client.Should(t, task.Create(&r))
func Should(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

// Check error and if present, fail and stop the test suite.
// Examples usage: client.Must(t, task.Create(&r))
func Must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
