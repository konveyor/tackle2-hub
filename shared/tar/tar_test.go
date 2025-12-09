package tar

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/onsi/gomega"
)

func TestWriter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Setup
	tmpDir, err := os.MkdirTemp("", "tar-*")
	g.Expect(err).To(gomega.BeNil())
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()
	outPath := path.Join(tmpDir, "output.tar.gz")
	file, err := os.Create(outPath)
	g.Expect(err).To(gomega.BeNil())

	// Write the ./data tree.
	writer := NewWriter(file)
	err = writer.AddDir("./data")
	g.Expect(err).To(gomega.BeNil())

	// Write ./data/rabbit => data/pet/rabbit
	err = writer.AddFile("./data/rabbit", "data/pet/rabbit")
	g.Expect(err).To(gomega.BeNil())
	writer.Close()
	_ = file.Close()

	// Read/expand the tarball.
	reader := NewReader()
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
				_ = equalFileContent(p, path.Join(tmpDir, p))
			}
			return
		})
	g.Expect(err).To(gomega.BeNil())

	// Validate ./data/pet/rabbit
	_ = equalFileContent("./data/rabbit", path.Join(tmpDir, "data", "pet", "rabbit"))
}

func equalFileContent(gotPath, expectedPath string) bool {
	got, err := os.Open(gotPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = got.Close()
	}()
	expected, err := os.Open(expectedPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = expected.Close()
	}()
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
