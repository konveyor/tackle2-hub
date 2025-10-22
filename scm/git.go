package scm

import (
	"errors"
	"fmt"
	urllib "net/url"
	"os"
	pathlib "path"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/command"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/konveyor/tackle2-hub/ssh"
)

// Git repository.
type Git struct {
	Base
}

// Validate settings.
func (r *Git) Validate() (err error) {
	u := GitURL{}
	err = u.With(r.Remote.URL)
	if err != nil {
		return
	}
	switch u.Scheme {
	case "http":
		if !r.Insecure {
			err = errors.New("http URL used with git.insecure.enabled = FALSE")
			return
		}
	}
	return
}

// Fetch clones the repository.
func (r *Git) Fetch() (err error) {
	err = nas.MkDir(r.Home(), 0755)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.writeConfig()
	if err != nil {
		return
	}
	err = r.writeCreds()
	if err != nil {
		return
	}
	u := r.URL()
	agent := ssh.New(r.Home())
	err = agent.Add(&r.Identity, u.Host)
	if err != nil {
		return
	}
	cmd := r.git()
	cmd.Options.Add("clone")
	cmd.Options.Add("--depth", "1")
	if r.Remote.Branch != "" {
		cmd.Options.Add("--single-branch")
		cmd.Options.Add("--branch", r.Remote.Branch)
	}
	cmd.Options.Add(u.String(), r.Path)
	err = cmd.Run()
	if err != nil {
		return
	}
	err = r.checkout()
	return
}

// Branch creates a branch with the given name if not exist and switch to it.
func (r *Git) Branch(ref string) (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("checkout", ref)
	err = cmd.Run()
	if err != nil {
		cmd = command.New("/usr/bin/git")
		cmd.Dir = r.Path
		cmd.Options.Add("checkout", "-b", ref)
	}
	r.Remote.Branch = ref
	err = cmd.Run()
	return
}

// addFiles adds files to staging area.
func (r *Git) addFiles(files []string) (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("add", files...)
	err = cmd.Run()
	return
}

// Commit files and push to remote.
func (r *Git) Commit(files []string, msg string) (err error) {
	err = r.addFiles(files)
	if err != nil {
		return err
	}
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("commit")
	cmd.Options.Add("--allow-empty")
	cmd.Options.Add("-m", msg)
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = r.push()
	return
}

// Head returns HEAD commit.
func (r *Git) Head() (commit string, err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("rev-parse")
	cmd.Options.Add("HEAD")
	err = cmd.Run()
	if err != nil {
		return
	}
	commit = string(cmd.Output())
	commit = strings.TrimSpace(commit)
	return
}

// URL returns the parsed URL.
func (r *Git) URL() (u GitURL) {
	u = GitURL{}
	_ = u.With(r.Remote.URL)
	return
}

// git returns git command.
func (r *Git) git() (cmd *command.Command) {
	cmd = command.New("/usr/bin/git")
	cmd.Env = append(
		os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"HOME="+r.Home())
	return
}

// push changes to remote.
func (r *Git) push() (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("push", "origin", "HEAD")
	err = cmd.Run()
	return
}

// writeConfig writes config file.
func (r *Git) writeConfig() (err error) {
	path := pathlib.Join(r.Home(), ".gitconfig")
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
	s := "[user]\n"
	s += "name = Konveyor Dev\n"
	s += "email = konveyor-dev@googlegroups.com\n"
	s += "[credential]\n"
	s += "helper = store --file="
	s += pathlib.Join(r.Home(), ".git-credentials")
	s += "\n"
	s += "[http]\n"
	s += fmt.Sprintf("sslVerify = %t\n", !r.Insecure)
	if proxy != "" {
		s += fmt.Sprintf("proxy = %s\n", proxy)
	}
	_, err = f.Write([]byte(s))
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

// writeCreds writes credentials (store) file.
func (r *Git) writeCreds() (err error) {
	if r.Identity.User == "" || r.Identity.Password == "" {
		return
	}
	Log.Info(
		fmt.Sprintf("Using identity: (id=%d) %s",
			r.Identity.ID,
			r.Identity.Name))
	path := pathlib.Join(r.Home(), ".git-credentials")
	f, err := os.Create(path)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	url := r.URL()
	for _, scheme := range []string{
		"https",
		"http",
	} {
		entry := scheme
		entry += "://"
		if r.Identity.User != "" {
			entry += urllib.PathEscape(r.Identity.User)
			entry += ":"
		}
		if r.Identity.Password != "" {
			entry += urllib.PathEscape(r.Identity.Password)
			entry += "@"
		}
		entry += url.Host
		_, err = f.Write([]byte(entry + "\n"))
		if err != nil {
			err = liberr.Wrap(
				err,
				"path",
				path)
			break
		}
	}
	_ = f.Close()
	Log.Info("Created: " + path)
	return
}

// proxy builds the proxy.
func (r *Git) proxy() (proxy string, err error) {
	kind := ""
	url := r.URL()
	switch url.Scheme {
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
		if h == url.Host {
			return
		}
	}
	Log.Info(
		fmt.Sprintf("Using proxy:(id=%d) %s",
			p.ID,
			p.Kind))
	auth := ""
	if p.Identity != nil {
		id := p.Identity
		auth = fmt.Sprintf(
			"%s:%s@",
			id.User,
			id.Password)
	}
	proxy = fmt.Sprintf(
		"http://%s%s",
		auth,
		p.Host)
	if p.Port > 0 {
		proxy = fmt.Sprintf(
			"%s:%d",
			proxy,
			p.Port)
	}
	return
}

// checkout ref.
func (r *Git) checkout() (err error) {
	branch := r.Remote.Branch
	if branch == "" {
		return
	}
	dir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(dir)
	}()
	_ = os.Chdir(r.Path)
	cmd := r.git()
	cmd.Options.Add("checkout", branch)
	err = cmd.Run()
	return
}

// GitURL git clone URL.
type GitURL struct {
	Raw    string
	Scheme string
	Host   string
	Path   string
}

// With populates the URL.
func (r *GitURL) With(u string) (err error) {
	r.Raw = u
	parsed, pErr := urllib.Parse(u)
	if pErr == nil {
		r.Scheme = parsed.Scheme
		r.Host = parsed.Host
		r.Path = parsed.Path
		return
	}
	notValid := liberr.New("URL not valid.")
	part := strings.Split(u, ":")
	if len(part) != 2 {
		err = notValid
		return
	}
	r.Host = part[0]
	r.Path = part[1]
	part = strings.Split(r.Host, "@")
	if len(part) != 2 {
		err = notValid
		return
	}
	r.Host = part[1]
	return
}

// String representation.
func (r *GitURL) String() string {
	return r.Raw
}
