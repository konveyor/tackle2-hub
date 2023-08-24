package tar

import (
	uuid2 "github.com/google/uuid"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/konveyor/tackle2-hub/tar"
	"github.com/konveyor/tackle2-hub/test/assert"
	"github.com/onsi/gomega"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestWriter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Setup
	uuid, _ := uuid2.NewUUID()
	tmpDir := "/tmp/" + uuid.String()
	err := nas.MkDir(tmpDir, 0755)
	g.Expect(err).To(gomega.BeNil())
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()
	outPath := path.Join(tmpDir, "output.tar.gz")
	file, err := os.Create(outPath)
	g.Expect(err).To(gomega.BeNil())

	// Write the ./data tree.
	writer := tar.NewWriter(file)
	err = writer.AddDir("./data")
	g.Expect(err).To(gomega.BeNil())

	// Write ./data/rabbit => data/pet/rabbit
	err = writer.AddFile("./data/rabbit", "data/pet/rabbit")
	g.Expect(err).To(gomega.BeNil())
	writer.Close()
	_ = file.Close()

	// Read/expand the tarball.
	reader := tar.NewReader()
	file, err = os.Open(outPath)
	g.Expect(err).To(gomega.BeNil())
	err = reader.Extract(tmpDir, file)
	g.Expect(err).To(gomega.BeNil())

	// Validate ./data
	err = filepath.Walk(
		path.Join("./data"),
		func(p string, info os.FileInfo, nErr error) (err error) {
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			if !info.IsDir() {
				_ = assert.EqualFileContent(p, path.Join(tmpDir, p))
			}
			return
		})
	g.Expect(err).To(gomega.BeNil())

	// Validate ./data/pet/rabbit
	_ = assert.EqualFileContent("./data/rabbit", path.Join(tmpDir, "data", "pet", "rabbit"))
}
