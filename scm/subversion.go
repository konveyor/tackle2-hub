package scm

import (
	"errors"
	"fmt"
	"io"
	urllib "net/url"
	"os"
	pathlib "path"
	"regexp"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/command"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/konveyor/tackle2-hub/ssh"
)

// RevisionRegex match revision embedded in `svn info`.
var RevisionRegex = regexp.MustCompile(`(Revision\:\s+)([\d]+)`)

// Subversion repository.
type Subversion struct {
	Base
}

// Validate settings.
func (r *Subversion) Validate() (err error) {
	u := SvnURL{}
	err = u.With(r.Remote)
	if err != nil {
		return
	}
	switch u.Scheme {
	case "http":
		if !r.Insecure {
			err = errors.New("http URL used with snv.insecure.enabled = FALSE")
			return
		}
	}
	return
}

// Fetch clones the repository.
func (r *Subversion) Fetch() (err error) {
	err = nas.MkDir(r.Home, 0755)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.writeConfig()
	if err != nil {
		return
	}
	err = r.writePassword()
	if err != nil {
		return
	}
	agent := ssh.Agent{}
	u := r.URL()
	err = agent.Add(&r.Identity, u.Host)
	if err != nil {
		return
	}
	err = r.checkout()
	return
}

// Branch checks out a branch.
// The branch is created as needed. The `ref` may be either:
// - fully qualified URL (includes branch and root path)
// - branch|tag path. (branches/stable).
func (r *Subversion) Branch(ref string) (err error) {
	branch := Subversion{}
	branch.Authenticated = r.Authenticated
	branch.Remote = r.Remote
	branch.Path = r.Path
	_, err = urllib.Parse(ref)
	if err == nil {
		branch.Remote = Remote{URL: ref}
	} else {
		branch.Remote.Branch = ref
	}
	branch.Remote = Remote{URL: ref}
	defer func() {
		if err == nil {
			r.Remote.URL = branch.Remote.URL
		}
	}()
	err = branch.checkout()
	if err != nil {
		err = branch.createBranch(r.Remote.URL)
	}
	return
}

// Commit records changes to the repo and push to the server
func (r *Subversion) Commit(files []string, msg string) (err error) {
	err = r.addFiles(files)
	if err != nil {
		return
	}
	cmd := r.svn()
	cmd.Dir = r.root()
	cmd.Options.Add("commit", "-m", msg)
	err = cmd.Run()
	return
}

// Head returns the current revision.
func (r *Subversion) Head() (commit string, err error) {
	cmd := r.svn()
	cmd.Dir = r.root()
	cmd.Options.Add("info", "-r", "HEAD")
	err = cmd.Run()
	if err != nil {
		return
	}
	output := cmd.Output()
	m := RevisionRegex.FindStringSubmatch(string(output))
	if len(m) == 3 {
		commit = m[2]
	} else {
		err = liberr.New("[SVN] info parser failed.")
	}
	return
}

// URL returns the parsed URL.
func (r *Subversion) URL() (u *SvnURL) {
	u = &SvnURL{}
	_ = u.With(r.Remote)
	return
}

// svn returns an svn command.
func (r *Subversion) svn() (cmd *command.Command) {
	cmd = command.New("/usr/bin/svn")
	cmd.Env = append(os.Environ(), "HOME="+r.Home)
	cmd.Options.Add("--non-interactive")
	if r.Insecure {
		cmd.Options.Add("--trust-server-cert")
	}
	return
}

// root returns a path to the cloned repository.
func (r *Subversion) root() (p string) {
	p = pathlib.Join(r.Path, r.Remote.Path)
	return
}

// checkout the repository.
func (r *Subversion) checkout() (err error) {
	root := r.root()
	_ = nas.RmDir(r.Path)
	_ = nas.MkDir(root, 0777)
	u := r.URL()
	cmd := r.svn()
	cmd.Options.Add("checkout", u.String(), root)
	err = cmd.Run()
	return
}

// createBranch create and checkout a branch.
func (r *Subversion) createBranch(baseURL string) (err error) {
	u := r.URL()
	cmd := r.svn()
	cmd.Options.Add(
		"copy",
		baseURL,
		u.String(),
		"-m",
		"Create branch: "+u.String())
	err = cmd.Run()
	if err != nil {
		return
	}
	err = r.checkout()
	return
}

// addFiles adds files to staging area
func (r *Subversion) addFiles(files []string) (err error) {
	cmd := r.svn()
	cmd.Dir = r.root()
	cmd.Options.Add("add")
	cmd.Options.Add("--force", files...)
	err = cmd.Run()
	return
}

// writeConfig writes configuration file.
func (r *Subversion) writeConfig() (err error) {
	path := pathlib.Join(
		r.Home,
		".subversion",
		"servers")
	err = nas.MkDir(pathlib.Dir(path), 0755)
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	proxy, err := r.proxy()
	if err != nil {
		return
	}
	_, err = f.Write([]byte(proxy))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	_ = f.Close()
	Log.Info("Created: " + path)
	return
}

// writePassword injects the password into: auth/svn.simple.
func (r *Subversion) writePassword() (err error) {
	if r.Identity.User == "" || r.Identity.Password == "" {
		return
	}
	Log.Info(
		fmt.Sprintf("Using identity:(id=%d) %s",
			r.Identity.ID,
			r.Identity.Name))
	cmd := r.svn()
	cmd.Options.Add("--username")
	cmd.Options.Add(r.Identity.User)
	cmd.Options.Add("--password")
	cmd.Options.Add(r.Identity.Password)
	cmd.Options.Add("info", r.URL().String())
	err = cmd.Run()
	if err != nil {
		return
	}
	dir := pathlib.Join(
		r.Home,
		".subversion",
		"auth",
		"svn.simple")
	files, err := os.ReadDir(dir)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			dir)
		return
	}
	path := pathlib.Join(dir, files[0].Name())
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	content, err := io.ReadAll(f)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	s := "K 8\n"
	s += "passtype\n"
	s += "V 6\n"
	s += "simple\n"
	s += "K 8\n"
	s += "username\n"
	s += fmt.Sprintf("V %d\n", len(r.Identity.User))
	s += fmt.Sprintf("%s\n", r.Identity.User)
	s += "K 8\n"
	s += "password\n"
	s += fmt.Sprintf("V %d\n", len(r.Identity.Password))
	s += fmt.Sprintf("%s\n", r.Identity.Password)
	s += string(content)
	_, err = f.Write([]byte(s))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	Log.Info("Created: " + path)
	return
}

// proxy builds the proxy.
func (r *Subversion) proxy() (proxy string, err error) {
	kind := ""
	u := r.URL()
	switch u.Scheme {
	case "http":
		kind = "http"
	case "https",
		"git@github.com":
		kind = "https"
	default:
		return
	}
	p, found := r.Proxies[kind]
	if !found || !p.Enabled {
		return
	}
	for _, h := range p.Excluded {
		if h == u.Host {
			return
		}
	}
	Log.Info(
		fmt.Sprintf("Using proxy:(id=%d) %s",
			p.ID,
			p.Kind))
	proxy = "[global]\n"
	proxy += fmt.Sprintf("http-proxy-host = %s\n", p.Host)
	if p.Port > 0 {
		proxy += fmt.Sprintf("http-proxy-port = %d\n", p.Port)
	}
	id := p.Identity
	if id != nil {
		proxy += fmt.Sprintf("http-proxy-username = %s\n", id.User)
		proxy += fmt.Sprintf("http-proxy-password = %s\n", id.Password)
	}
	proxy += fmt.Sprintf(
		"(http-proxy-exceptions = %s\n",
		strings.Join(p.Excluded, " "))
	return
}

// SvnURL subversion URL.
type SvnURL struct {
	Raw      string
	Branch   string
	RootPath string
	Scheme   string
	Host     string
}

// With initializes with a remote.
func (u *SvnURL) With(r Remote) (err error) {
	parsed, err := urllib.Parse(r.URL)
	if err != nil {
		return
	}
	u.Raw = r.URL
	u.Branch = r.Branch
	u.RootPath = r.Path
	u.Scheme = parsed.Scheme
	u.Host = parsed.Host
	return
}

// String returns a URL with Branch and RootPath appended.
func (u *SvnURL) String() (s string) {
	parsed, _ := urllib.Parse(u.Raw)
	parsed.Path = pathlib.Join(parsed.Path, u.Branch, u.RootPath)
	s = parsed.String()
	return
}
