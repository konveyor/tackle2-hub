package assert

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Compare two files content using sha256sum function.
func EqualFileContent(gotPath, expectedPath string) bool {
	got, err := os.Open(gotPath)
	if err != nil {
		panic(err)
	}
	expected, err := os.Open(expectedPath)
	if err != nil {
		panic(err)
	}
	defer got.Close()
	defer expected.Close()

	hGot := sha256.New()
	if _, err := io.Copy(hGot, got); err != nil {
		panic(err)
	}
	hExpected := sha256.New()
	if _, err := io.Copy(hExpected, expected); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%v", hGot.Sum(nil)) == fmt.Sprintf("%v", hExpected.Sum(nil))
}
