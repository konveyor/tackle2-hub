package binding

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a temporary directory for test files
	testDir, err := os.MkdirTemp("", "test-bucket-*")
	g.Expect(err).To(BeNil())

	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Create test files in the temp directory
	testFile := filepath.Join(testDir, "file1.txt")
	err = os.WriteFile(testFile, []byte("Hello from file1"), 0644)
	g.Expect(err).To(BeNil())

	testFile2 := filepath.Join(testDir, "file2.txt")
	err = os.WriteFile(testFile2, []byte("Hello from file2"), 0644)
	g.Expect(err).To(BeNil())

	// Create subdirectory with a file
	subDir := filepath.Join(testDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	g.Expect(err).To(BeNil())

	testFile3 := filepath.Join(subDir, "file3.txt")
	err = os.WriteFile(testFile3, []byte("Hello from subdir"), 0644)
	g.Expect(err).To(BeNil())

	// Define the bucket to create
	bucket := &api.Bucket{
		Path: "test-bucket-" + t.Name(),
	}

	// CREATE: Create the bucket
	err = client.Bucket.Create(bucket)
	g.Expect(err).To(BeNil())
	g.Expect(bucket.ID).NotTo(BeZero())

	defer func() {
		_ = client.Bucket.Delete(bucket.ID)
	}()

	// GET: List buckets
	list, err := client.Bucket.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(bucket, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the bucket and verify it matches
	retrieved, err := client.Bucket.Get(bucket.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(bucket, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// CONTENT OPERATIONS: Test file upload and download
	selected := client.Bucket.Select(bucket.ID)

	// Upload a single file
	err = selected.Content.Put(testFile, "uploaded-file1.txt")
	g.Expect(err).To(BeNil())

	// Download the file and verify content
	downloadedFile, err := os.CreateTemp("", "test-downloaded-*")
	g.Expect(err).To(BeNil())
	_ = downloadedFile.Close()

	defer func() {
		_ = os.Remove(downloadedFile.Name())
	}()

	err = selected.Content.Get("uploaded-file1.txt", downloadedFile.Name())
	g.Expect(err).To(BeNil())

	// Verify the downloaded file matches the original
	g.Expect(assert.EqualFileContent(downloadedFile.Name(), testFile)).To(BeTrue())

	// Upload another file
	err = selected.Content.Put(testFile2, "uploaded-file2.txt")
	g.Expect(err).To(BeNil())

	// Upload a directory
	err = selected.Content.Put(testDir, "uploaded-dir")
	g.Expect(err).To(BeNil())

	// Download the directory and verify content
	downloadedDir, err := os.MkdirTemp("", "test-downloaded-dir-*")
	g.Expect(err).To(BeNil())

	defer func() {
		_ = os.RemoveAll(downloadedDir)
	}()

	err = selected.Content.Get("uploaded-dir", downloadedDir)
	g.Expect(err).To(BeNil())

	// Verify directory structure
	// Check that the files exist in the downloaded directory
	downloadedFile1 := filepath.Join(downloadedDir, "file1.txt")
	g.Expect(assert.EqualFileContent(downloadedFile1, testFile)).To(BeTrue())

	downloadedFile2 := filepath.Join(downloadedDir, "file2.txt")
	g.Expect(assert.EqualFileContent(downloadedFile2, testFile2)).To(BeTrue())

	downloadedFile3 := filepath.Join(downloadedDir, "subdir", "file3.txt")
	g.Expect(assert.EqualFileContent(downloadedFile3, testFile3)).To(BeTrue())

	// DELETE CONTENT: Remove a file from the bucket
	err = selected.Content.Delete("uploaded-file1.txt")
	g.Expect(err).To(BeNil())

	// Verify the file was deleted (download should fail)
	deletedFile, err := os.CreateTemp("", "test-deleted-*")
	g.Expect(err).To(BeNil())
	deletedFilePath := deletedFile.Name()
	_ = deletedFile.Close()

	defer func() {
		_ = os.Remove(deletedFilePath)
	}()

	err = selected.Content.Get("uploaded-file1.txt", deletedFilePath)
	g.Expect(err).ToNot(BeNil())

	// DELETE CONTENT: Remove a directory from the bucket
	err = selected.Content.Delete("uploaded-dir")
	g.Expect(err).To(BeNil())

	// Verify the directory was deleted (download should fail)
	deletedDir, err := os.MkdirTemp("", "test-deleted-dir-*")
	g.Expect(err).To(BeNil())

	defer func() {
		_ = os.RemoveAll(deletedDir)
	}()

	err = selected.Content.Get("uploaded-dir", deletedDir)
	g.Expect(err).ToNot(BeNil())

	// DELETE: Remove the bucket
	err = client.Bucket.Delete(bucket.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Bucket.Get(bucket.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
