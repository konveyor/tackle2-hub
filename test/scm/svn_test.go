package scm

import (
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/scm"
	"github.com/onsi/gomega"
)

const (
	EnvSvnURL      = "SVN_URL"
	EnvSvnUser     = "SVN_USER"
	EnvSvnPassword = "SVN_PASSWORD"
)

// newSvn builds an scm.SCM for the svn repo defined by env vars.
func newSvn(t *testing.T) (r scm.SCM) {
	t.Helper()
	r = newSvnBranch(t, "")
	return
}

// newSvnBranch builds an scm.SCM for the svn repo, checked out
// at the given base branch (e.g. "trunk").
func newSvnBranch(t *testing.T, branch string) (r scm.SCM) {
	t.Helper()
	svnURL := os.Getenv(EnvSvnURL)
	user := os.Getenv(EnvSvnUser)
	password := os.Getenv(EnvSvnPassword)
	if svnURL == "" || user == "" || password == "" {
		t.Skip("SVN_URL, SVN_USER and SVN_PASSWORD required.")
	}
	destDir := t.TempDir()
	repository := api.Repository{
		Kind:   "svn",
		URL:    svnURL,
		Branch: branch,
	}
	identity := &scm.Identity{
		User:     user,
		Password: password,
	}
	var err error
	r = scm.New(destDir, repository)
	err = r.Validate()
	if err != nil {
		t.Fatalf("scm.Validate() failed: %v", err)
	}
	r.WithIdentity(identity)
	t.Cleanup(func() {
		_ = r.Clean()
	})
	return
}

// svnBranchName returns a unique branch path.
func svnBranchName() (name string) {
	name = fmt.Sprintf("branches/test-%d", time.Now().UnixMilli())
	return
}

// deleteSvnBranch removes a branch path from the remote for cleanup.
func deleteSvnBranch(r scm.SCM, branch string) (err error) {
	svn, cast := r.(*scm.Subversion)
	if !cast {
		return
	}
	identity := svn.Remote.Identity
	cmd := osexec.Command(
		"svn",
		"remove",
		svn.Remote.URL+"/"+branch,
		"-m", "cleanup",
		"--username", identity.User,
		"--password", identity.Password,
		"--non-interactive")
	err = cmd.Run()
	return
}

// TestSvnId verifies that Id returns a stable, non-empty digest.
func TestSvnId(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	id := r.Id()
	g.Expect(id).NotTo(gomega.BeEmpty())

	// Stable across calls.
	g.Expect(r.Id()).To(gomega.Equal(id))
}

// TestSvnValidate verifies that Validate succeeds for a valid remote.
func TestSvnValidate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Validate()
	g.Expect(err).To(gomega.BeNil())
}

// TestSvnFetch verifies that Fetch checks out the repository.
func TestSvnFetch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestSvnUpdate verifies that Update pulls the latest changes.
func TestSvnUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Update()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestSvnCommit verifies staging, committing, and pushing a file.
func TestSvnCommit(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	before, err := r.Head()
	g.Expect(err).To(gomega.BeNil())

	svn := r.(*scm.Subversion)
	name := fmt.Sprintf("test-%d.txt", time.Now().UnixMilli())
	dest := filepath.Join(svn.Path, name)
	err = os.WriteFile(dest, []byte("test-content"), 0644)
	g.Expect(err).To(gomega.BeNil())

	err = r.Commit([]string{name}, "test commit")
	g.Expect(err).To(gomega.BeNil())

	after, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(after).NotTo(gomega.Equal(before))
}

// TestSvnPush verifies that Push succeeds (a no-op for svn).
func TestSvnPush(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Push()
	g.Expect(err).To(gomega.BeNil())
}

// TestSvnHead verifies that Head returns a non-empty revision.
func TestSvnHead(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestSvnClean verifies that Clean removes the home directory.
func TestSvnClean(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvn(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Clean()
	g.Expect(err).To(gomega.BeNil())
}

// TestSvnBranch verifies switching to an existing branch.
func TestSvnBranch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvnBranch(t, "trunk")

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Branch("trunk")
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestSvnBranchNotFound verifies that Branch without CREATE returns
// an error when the branch does not exist.
func TestSvnBranchNotFound(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvnBranch(t, "trunk")

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Branch("branches/does-not-exist")
	g.Expect(err).NotTo(gomega.BeNil())
}

// TestSvnBranchCreate verifies creating a new branch with CREATE.
func TestSvnBranchCreate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newSvnBranch(t, "trunk")

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	branch := svnBranchName()
	t.Cleanup(func() {
		_ = deleteSvnBranch(r, branch)
	})

	err = r.Branch(branch, scm.CREATE)
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}
