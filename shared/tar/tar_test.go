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

func TestFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	root := "/not/created"

	// No patterns matches everything.
	f := NewFilter(root)
	g.Expect(f.Match(path.Join(root, "a.yaml"))).To(gomega.BeTrue())

	// Included patterns match paths that have not been created.
	f = NewFilter(root)
	f.Include("*.yaml")
	g.Expect(f.Match(path.Join(root, "a.yaml"))).To(gomega.BeTrue())
	g.Expect(f.Match(path.Join(root, "b.txt"))).To(gomega.BeFalse())

	// Patterns are rooted.
	g.Expect(f.Match("/other/a.yaml")).To(gomega.BeFalse())

	// (*) does not cross separators.
	g.Expect(f.Match(path.Join(root, "sub", "c.yaml"))).To(gomega.BeFalse())
	f = NewFilter(root)
	f.Include("sub/*.yaml")
	g.Expect(f.Match(path.Join(root, "sub", "c.yaml"))).To(gomega.BeTrue())
	g.Expect(f.Match(path.Join(root, "c.yaml"))).To(gomega.BeFalse())

	// Excluded.
	f = NewFilter(root)
	f.Exclude("*.tmp")
	g.Expect(f.Match(path.Join(root, "a.yaml"))).To(gomega.BeTrue())
	g.Expect(f.Match(path.Join(root, "a.tmp"))).To(gomega.BeFalse())

	// Excluded takes precedence over included.
	f = NewFilter(root)
	f.Include("*.yaml", "*.tmp")
	f.Exclude("*.tmp")
	g.Expect(f.Match(path.Join(root, "a.yaml"))).To(gomega.BeTrue())
	g.Expect(f.Match(path.Join(root, "a.tmp"))).To(gomega.BeFalse())

	// Empty patterns ignored.
	f = NewFilter(root)
	f.Include("")
	f.Exclude("")
	g.Expect(f.Match(path.Join(root, "a.yaml"))).To(gomega.BeTrue())
}

func TestReaderFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	tmpDir, err := os.MkdirTemp("", "tar-*")
	g.Expect(err).To(gomega.BeNil())
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()

	// Write the ./data tree.
	outPath := path.Join(tmpDir, "output.tar.gz")
	file, err := os.Create(outPath)
	g.Expect(err).To(gomega.BeNil())
	writer := NewWriter(file)
	err = writer.AddDir("./data")
	g.Expect(err).To(gomega.BeNil())
	writer.Close()
	_ = file.Close()

	// Extract only data/pet.
	// The enclosing DIRECTORY entries must be included because
	// Extract creates directories only from (TypeDir) entries.
	outDir := path.Join(tmpDir, "out")
	filter := NewFilter(outDir)
	filter.Include("data", "data/pet", "data/pet/*")
	reader := NewReader()
	reader.Filter = filter
	file, err = os.Open(outPath)
	g.Expect(err).To(gomega.BeNil())
	defer func() {
		_ = file.Close()
	}()
	err = reader.Extract(outDir, file)
	g.Expect(err).To(gomega.BeNil())

	// Included entries extracted.
	for _, name := range []string{"data/pet/cat", "data/pet/dog"} {
		_, err = os.Stat(path.Join(outDir, name))
		g.Expect(err).To(gomega.BeNil(), name)
	}
	// Entries after the first filtered entry still extracted.
	// Excluded entries not extracted.
	for _, name := range []string{"data/rabbit", "data/wild/lion", "data/wild/tiger"} {
		_, err = os.Stat(path.Join(outDir, name))
		g.Expect(os.IsNotExist(err)).To(gomega.BeTrue(), name)
	}
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
