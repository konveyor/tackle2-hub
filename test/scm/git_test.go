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
	EnvGitURL   = "GIT_URL"
	EnvGitToken = "GIT_TOKEN"
)

// newGit builds an scm.SCM for the git repo defined by env vars.
func newGit(t *testing.T) (r scm.SCM) {
	t.Helper()
	gitURL := os.Getenv(EnvGitURL)
	gitToken := os.Getenv(EnvGitToken)
	if gitURL == "" || gitToken == "" {
		t.Skip("GIT_URL and GIT_TOKEN required.")
	}
	destDir := t.TempDir()
	repository := api.Repository{
		Kind: "git",
		URL:  gitURL,
	}
	identity := &scm.Identity{
		User:     "token",
		Password: gitToken,
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

// branchName returns a unique branch name.
func branchName() (name string) {
	name = fmt.Sprintf("test-%d", time.Now().UnixMilli())
	return
}

// TestId verifies that Id returns a stable, non-empty digest.
func TestId(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	id := r.Id()
	g.Expect(id).NotTo(gomega.BeEmpty())

	// Stable across calls.
	g.Expect(r.Id()).To(gomega.Equal(id))
}

// TestValidate verifies that Validate succeeds for a valid remote.
func TestValidate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Validate()
	g.Expect(err).To(gomega.BeNil())
}

// TestFetch verifies that Fetch clones the repository.
func TestFetch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestUpdate verifies that Update pulls the latest changes.
func TestUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Update()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestBranch verifies switching to an existing branch.
func TestBranch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Branch("main")
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestBranch verifies switching to the default branch.
func TestDefaultBranch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Branch("")
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestBranchNotFound verifies that Branch without CREATE returns
// an error when the branch does not exist.
func TestBranchNotFound(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Branch("branch-does-not-exist")
	g.Expect(err).NotTo(gomega.BeNil())
}

// TestBranchCreate verifies creating a new branch with CREATE.
func TestBranchCreate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	branch := branchName()
	t.Cleanup(func() {
		_ = deleteRemoteBranch(r, branch)
	})

	err = r.Branch(branch, scm.CREATE)
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
}

// TestCommit verifies staging, committing, and pushing a file.
func TestCommit(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	branch := branchName()
	t.Cleanup(func() {
		_ = deleteRemoteBranch(r, branch)
	})

	err = r.Branch(branch, scm.CREATE)
	g.Expect(err).To(gomega.BeNil())

	before, err := r.Head()
	g.Expect(err).To(gomega.BeNil())

	git := r.(*scm.Git)
	dest := filepath.Join(git.Path, "test.txt")
	err = os.WriteFile(dest, []byte("test-content"), 0644)
	g.Expect(err).To(gomega.BeNil())

	err = r.Commit([]string{"test.txt"}, "test commit")
	g.Expect(err).To(gomega.BeNil())

	after, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(after).NotTo(gomega.Equal(before))
}

// TestPush verifies pushing changes to the remote.
func TestPush(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	branch := branchName()
	t.Cleanup(func() {
		_ = deleteRemoteBranch(r, branch)
	})

	err = r.Branch(branch, scm.CREATE)
	g.Expect(err).To(gomega.BeNil())

	// Commit creates a commit (allow-empty) and pushes internally,
	// but Push() is the explicit push path.
	git := r.(*scm.Git)
	dest := filepath.Join(git.Path, "push-test.txt")
	err = os.WriteFile(dest, []byte("push-content"), 0644)
	g.Expect(err).To(gomega.BeNil())

	err = r.Commit([]string{"push-test.txt"}, "pre-push commit")
	g.Expect(err).To(gomega.BeNil())

	err = r.Push()
	g.Expect(err).To(gomega.BeNil())
}

// TestHead verifies that Head returns a non-empty commit hash.
func TestHead(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	commit, err := r.Head()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(commit).NotTo(gomega.BeEmpty())
	g.Expect(len(commit)).To(gomega.BeNumerically(">=", 7))
}

// TestClean verifies that Clean removes the home directory.
func TestClean(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := newGit(t)

	err := r.Fetch()
	g.Expect(err).To(gomega.BeNil())

	err = r.Clean()
	g.Expect(err).To(gomega.BeNil())
}

// deleteRemoteBranch deletes a remote branch for cleanup.
// Runs initHome-equivalent setup through Branch to ensure
// credentials are available, then deletes via the repo's
// configured working tree.
func deleteRemoteBranch(r scm.SCM, branch string) (err error) {
	git, cast := r.(*scm.Git)
	if !cast {
		return
	}
	cmd := osexec.Command(
		"git",
		"-C", git.Path,
		"push", "origin",
		"--delete", branch)
	cmd.Env = append(
		os.Environ(),
		"HOME="+git.Home)
	err = cmd.Run()
	return
}
