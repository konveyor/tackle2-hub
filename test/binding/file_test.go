package binding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestFile(t *testing.T) {
	g := NewGomegaWithT(t)

	ignoredPaths := []string{
		"CreateTime",
		"CreateUser",
		"UpdateUser",
		"Path",
		"Encoding",
	}

	// Create a temporary test file to upload
	testDir, err := os.MkdirTemp("", "test-file-*")
	g.Expect(err).To(BeNil())

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})

	sourceFile := filepath.Join(testDir, "testfile.txt")
	testContent := []byte("This is test file content\nLine 2\nLine 3\n")
	err = os.WriteFile(sourceFile, testContent, 0644)
	g.Expect(err).To(BeNil())

	// CREATE: Upload the file using Put
	uploaded, err := client.File.Put(sourceFile)
	g.Expect(err).To(BeNil())
	g.Expect(uploaded).NotTo(BeNil())
	g.Expect(uploaded.ID).NotTo(BeZero())

	expectedFile := &api.File{
		Resource: api.Resource{ID: uploaded.ID},
		Name:     "testfile.txt",
	}
	eq, report := cmp.Eq(expectedFile, uploaded, ignoredPaths...)
	g.Expect(eq).To(BeTrue(), report)

	t.Cleanup(func() {
		_ = client.File.Delete(uploaded.ID)
	})

	// GET: Download the file to a specific path and verify content matches
	downloadFile, err := os.CreateTemp("", "test-download-*")
	g.Expect(err).To(BeNil())
	downloadPath := downloadFile.Name()
	_ = downloadFile.Close()

	t.Cleanup(func() {
		_ = os.Remove(downloadPath)
	})

	err = client.File.Get(uploaded.ID, downloadPath)
	g.Expect(err).To(BeNil())

	// Verify the downloaded file matches the original
	g.Expect(assert.EqualFileContent(downloadPath, sourceFile)).To(BeTrue())

	// GET: Download the file to a directory
	downloadDir, err := os.MkdirTemp("", "test-download-dir-*")
	g.Expect(err).To(BeNil())

	t.Cleanup(func() {
		_ = os.RemoveAll(downloadDir)
	})

	err = client.File.Get(uploaded.ID, downloadDir)
	g.Expect(err).To(BeNil())

	// Verify file was downloaded with correct name in the directory
	downloadedInDir := filepath.Join(downloadDir, "testfile.txt")
	g.Expect(assert.EqualFileContent(downloadedInDir, sourceFile)).To(BeTrue())

	// PATCH: Append content to the file
	appendContent := []byte("Appended line 1\nAppended line 2\n")
	err = client.File.Patch(uploaded.ID, appendContent)
	g.Expect(err).To(BeNil())

	// GET: Download the patched file and verify appended content
	patchedFile, err := os.CreateTemp("", "test-patched-*")
	g.Expect(err).To(BeNil())
	patchedPath := patchedFile.Name()
	_ = patchedFile.Close()

	t.Cleanup(func() {
		_ = os.Remove(patchedPath)
	})

	err = client.File.Get(uploaded.ID, patchedPath)
	g.Expect(err).To(BeNil())

	// Verify the patched file contains the appended content
	patchedData, err := os.ReadFile(patchedPath)
	g.Expect(err).To(BeNil())

	expectedData := append(testContent, appendContent...)
	g.Expect(patchedData).To(Equal(expectedData))

	// DELETE: Remove the file
	err = client.File.Delete(uploaded.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	deletedFile, err := os.CreateTemp("", "test-deleted-*")
	g.Expect(err).To(BeNil())
	deletedPath := deletedFile.Name()
	_ = deletedFile.Close()

	t.Cleanup(func() {
		_ = os.Remove(deletedPath)
	})

	err = client.File.Get(uploaded.ID, deletedPath)
	g.Expect(err).ToNot(BeNil())

	// CREATE: Test Touch to create an empty file
	emptyFile, err := client.File.Touch("empty.txt")
	g.Expect(err).To(BeNil())
	g.Expect(emptyFile).NotTo(BeNil())
	g.Expect(emptyFile.ID).NotTo(BeZero())

	expectedEmpty := &api.File{
		Resource: api.Resource{ID: emptyFile.ID},
		Name:     "empty.txt",
	}
	eq, report = cmp.Eq(expectedEmpty, emptyFile, ignoredPaths...)
	g.Expect(eq).To(BeTrue(), report)

	t.Cleanup(func() {
		_ = client.File.Delete(emptyFile.ID)
	})

	// GET: Download the empty file and verify it's empty
	emptyDownload, err := os.CreateTemp("", "test-empty-*")
	g.Expect(err).To(BeNil())
	emptyDownloadPath := emptyDownload.Name()
	_ = emptyDownload.Close()

	t.Cleanup(func() {
		_ = os.Remove(emptyDownloadPath)
	})

	err = client.File.Get(emptyFile.ID, emptyDownloadPath)
	g.Expect(err).To(BeNil())

	emptyData, err := os.ReadFile(emptyDownloadPath)
	g.Expect(err).To(BeNil())
	g.Expect(len(emptyData)).To(Equal(0))

	// DELETE: Remove the empty file
	err = client.File.Delete(emptyFile.ID)
	g.Expect(err).To(BeNil())

	// CREATE: Test Post method
	sourceFile2 := filepath.Join(testDir, "testfile2.txt")
	testContent2 := []byte("Content for Post test\n")
	err = os.WriteFile(sourceFile2, testContent2, 0644)
	g.Expect(err).To(BeNil())

	posted, err := client.File.Post(sourceFile2)
	g.Expect(err).To(BeNil())
	g.Expect(posted).NotTo(BeNil())
	g.Expect(posted.ID).NotTo(BeZero())

	expectedPosted := &api.File{
		Resource: api.Resource{ID: posted.ID},
		Name:     "testfile2.txt",
	}
	eq, report = cmp.Eq(expectedPosted, posted, ignoredPaths...)
	g.Expect(eq).To(BeTrue(), report)

	t.Cleanup(func() {
		_ = client.File.Delete(posted.ID)
	})

	// GET: Verify the posted file content
	postedDownload, err := os.CreateTemp("", "test-posted-*")
	g.Expect(err).To(BeNil())
	postedDownloadPath := postedDownload.Name()
	_ = postedDownload.Close()

	t.Cleanup(func() {
		_ = os.Remove(postedDownloadPath)
	})

	err = client.File.Get(posted.ID, postedDownloadPath)
	g.Expect(err).To(BeNil())

	g.Expect(assert.EqualFileContent(postedDownloadPath, sourceFile2)).To(BeTrue())

	// DELETE: Remove the posted file
	err = client.File.Delete(posted.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion
	err = client.File.Get(posted.ID, postedDownloadPath)
	g.Expect(err).ToNot(BeNil())
}
