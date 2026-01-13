package scm

import (
	"errors"
	"fmt"
	urllib "net/url"
	"os"
	"path/filepath"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/command"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/konveyor/tackle2-hub/shared/ssh"
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
		if !r.Remote.Insecure {
			err = errors.New("http URL used with git.insecure.enabled = FALSE")
			return
		}
	}
	return
}

// Fetch clones the repository.
func (r *Git) Fetch() (err error) {
	err = r.mustEmptyDir(r.Path)
	if err != nil {
		return
	}
	err = r.initHome()
	if err != nil {
		return
	}
	u := r.URL()
	cmd := r.git()
	cmd.Options.Add("clone")
	cmd.Options.Add(u.String(), r.Path)
	err = cmd.Run()
	if err != nil {
		return
	}
	if r.Remote.Branch != "" {
		err = r.checkout(r.Remote.Branch)
		if err != nil {
			return
		}
	}
	return
}

// Update the repository using the remote.
func (r *Git) Update() (err error) {
	err = r.initHome()
	if err != nil {
		return
	}
	err = r.fetch()
	if err != nil {
		return
	}
	err = r.pull()
	if err != nil {
		return
	}
	return
}

// Branch creates a branch with the given name if not exist and switch to it.
func (r *Git) Branch(ref string) (err error) {
	err = r.initHome()
	if err != nil {
		return
	}
	err = r.checkout(ref)
	return
}

// Commit files and push to remote.
func (r *Git) Commit(files []string, msg string) (err error) {
	err = r.initHome()
	if err != nil {
		return
	}
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
	err = r.initHome()
	if err != nil {
		return
	}
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

// initHome ensures the home directory is updated.
func (r *Git) initHome() (err error) {
	err = nas.RmDir(r.Home)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = nas.MkDir(r.Home, 0755)
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
	identity := r.Remote.Identity
	if identity != nil {
		key := ssh.Key{
			ID:         identity.ID,
			Name:       identity.Name,
			Content:    identity.Key,
			Passphrase: identity.Password,
		}
		Log.V(1).Info(
			fmt.Sprintf("[GIT] Using identity: (id=%d) %s",
				identity.ID,
				identity.Name))
		err = key.Add()
		if err != nil {
			return
		}
	}
	return
}

// git returns git command.
func (r *Git) git() (cmd *command.Command) {
	cmd = command.New("/usr/bin/git")
	cmd.Env = append(
		os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"HOME="+r.Home)
	return
}

// fetch refs and commits.
func (r *Git) fetch() (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("fetch", "--tags", "--prune")
	err = cmd.Run()
	return
}

// pull commits.
func (r *Git) pull() (err error) {
	isTag, err := r.isTag(r.Remote.Branch)
	if err != nil {
		return
	}
	if isTag {
		return
	}
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("pull")
	err = cmd.Run()
	return
}

// checkout ref.
func (r *Git) checkout(ref string) (err error) {
	if ref == "" {
		ref, err = r.defaultBranch()
		if err != nil {
			return
		}
	}
	defer func() {
		if err == nil {
			r.Remote.Branch = ref
		}
	}()
	err = r.fetch()
	if err != nil {
		return
	}
	isTag, err := r.isTag(ref)
	if err != nil {
		return
	}
	if isTag {
		cmd := r.git()
		cmd.Dir = r.Path
		cmd.Options.Add("checkout", ref)
		err = cmd.Run()
		return
	}
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("checkout", "-B", ref, "origin/"+ref)
	err = cmd.Run()
	if err != nil {
		cmd = r.git()
		cmd.Dir = r.Path
		cmd.Options.Add("checkout", ref)
		err = cmd.Run()
		if err != nil {
			return
		}
	}
	err = r.pull()
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

// push changes to remote.
func (r *Git) push() (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("push", "origin", "HEAD")
	err = cmd.Run()
	return
}

// defaultBranch returns the branch name.
func (r *Git) defaultBranch() (name string, err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("rev-parse", "--abbrev-ref", "origin/HEAD")
	err = cmd.Run()
	if err != nil {
		return
	}
	name = string(cmd.Output())
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	return
}

// isTag returns true when a when ref is a tag.
func (r *Git) isTag(ref string) (matched bool, err error) {
	if ref == "" {
		return
	}
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("tag")
	err = cmd.Run()
	if err != nil {
		return
	}
	output := string(cmd.Output())
	tags := strings.Split(output, "\n")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == ref {
			matched = true
			break
		}
	}
	return
}

// writeConfig writes config file.
func (r *Git) writeConfig() (err error) {
	path := filepath.Join(r.Home, ".gitconfig")
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
	s += filepath.Join(r.Home, ".git-credentials")
	s += "\n"
	s += "[http]\n"
	s += fmt.Sprintf("sslVerify = %t\n", !r.Remote.Insecure)
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
	Log.V(1).Info("[GIT] Created: " + path)
	return
}

// writeCreds writes credentials (store) file.
func (r *Git) writeCreds() (err error) {
	identity := r.Remote.Identity
	if identity == nil {
		return
	}
	if identity.User == "" || identity.Password == "" {
		return
	}
	Log.V(1).Info(
		fmt.Sprintf("[GIT] Using identity: (id=%d) %s",
			identity.ID,
			identity.Name))
	path := filepath.Join(r.Home, ".git-credentials")
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
		if identity.User != "" {
			entry += urllib.PathEscape(identity.User)
			entry += ":"
		}
		if identity.Password != "" {
			entry += urllib.PathEscape(identity.Password)
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
	Log.V(1).Info("[GIT] Created: " + path)
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
	if !found {
		return
	}
	for _, h := range p.Excluded {
		if h == url.Host {
			return
		}
	}
	Log.V(1).Info(
		fmt.Sprintf("[GIT] Using proxy:(id=%d) %s",
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
