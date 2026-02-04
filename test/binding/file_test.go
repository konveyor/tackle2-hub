package binding

import (
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
	. "github.com/onsi/gomega"
)

func TestFile(t *testing.T) {
	g := NewGomegaWithT(t)

	// Use /etc/hosts as the test file to upload
	sourceFile := "/etc/hosts"

	// PUT: Upload the file
	uploaded, err := client.File.Put(sourceFile)
	g.Expect(err).To(BeNil())
	g.Expect(uploaded).NotTo(BeNil())
	g.Expect(uploaded.ID).NotTo(BeZero())
	g.Expect(uploaded.Name).To(Equal("hosts"))

	defer func() {
		_ = client.File.Delete(uploaded.ID)
	}()

	// GET: Download the file and verify content matches
	downloadPath := "/tmp/test-hosts-download"
	err = client.File.Get(uploaded.ID, downloadPath)
	g.Expect(err).To(BeNil())

	defer func() {
		_ = os.Remove(downloadPath)
	}()

	// Verify the downloaded file matches the original
	g.Expect(assert.EqualFileContent(downloadPath, sourceFile)).To(BeTrue())

	// PATCH: Append content to the file
	appendContent := []byte("\n# Test comment added by file_test.go\n")
	err = client.File.Patch(uploaded.ID, appendContent)
	g.Expect(err).To(BeNil())

	// GET: Download the patched file and verify appended content
	patchedPath := "/tmp/test-hosts-patched"
	err = client.File.Get(uploaded.ID, patchedPath)
	g.Expect(err).To(BeNil())

	defer func() {
		_ = os.Remove(patchedPath)
	}()

	// Verify the patched file contains the appended content
	patchedData, err := os.ReadFile(patchedPath)
	g.Expect(err).To(BeNil())

	originalData, err := os.ReadFile(sourceFile)
	g.Expect(err).To(BeNil())

	expectedData := append(originalData, appendContent...)
	g.Expect(patchedData).To(Equal(expectedData))

	// DELETE: Remove the file
	err = client.File.Delete(uploaded.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	err = client.File.Get(uploaded.ID, "/dev/null")
	g.Expect(err).ToNot(BeNil())
}
