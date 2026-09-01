package command

import (
	"errors"
	"io"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNew(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	g.Expect(cmd).NotTo(BeNil())
	g.Expect(cmd.Path).To(Equal("/bin/echo"))
}

func TestOptionsAdd(t *testing.T) {
	g := NewGomegaWithT(t)
	options := Options{}
	options.Add("run")
	options.Add("--flag", "a", "b")
	g.Expect(options).To(Equal(Options{"run", "--flag", "a", "b"}))
}

func TestOptionsAddf(t *testing.T) {
	g := NewGomegaWithT(t)
	options := Options{}
	options.Addf("count=%d", 5)
	options.Addf("%s=%s", "name", "elmer")
	g.Expect(options).To(Equal(Options{"count=5", "name=elmer"}))
}

func TestSetExpandsOptions(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("--user", "${user}")
	cmd.Options.Add("--password", "${password}")
	cmd.Set("user", "elmer")
	cmd.Set("password", "secret")
	g.Expect(cmd.options()).To(
		Equal([]string{
			"--user", "elmer",
			"--password", "secret"}))
}

func TestExpandLeavesUnknownRef(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("--user", "${user}")
	cmd.Options.Add("--other", "${missing}")
	cmd.Set("user", "elmer")
	g.Expect(cmd.options()).To(
		Equal([]string{
			"--user", "elmer",
			"--other", "${missing}"}))
}

func TestExpandWithoutParams(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("--user", "${user}")
	g.Expect(cmd.options()).To(Equal([]string{"--user", "${user}"}))
}

func TestExpandDoesNotLeakToCommandString(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("--password", "${password}")
	cmd.Set("password", "secret")
	g.Expect(cmd.Command()).To(ContainSubstring("${password}"))
	g.Expect(cmd.Command()).NotTo(ContainSubstring("secret"))
}

func TestCommandString(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("hello", "world")
	g.Expect(cmd.Command()).To(Equal("/bin/echo hello world"))
}

func TestRun(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("hello")
	err := cmd.Run()
	g.Expect(err).To(BeNil())
	g.Expect(string(cmd.Output())).To(Equal("hello\n"))
}

func TestRunExpandsParams(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("${msg}")
	cmd.Set("msg", "hello")
	err := cmd.Run()
	g.Expect(err).To(BeNil())
	g.Expect(string(cmd.Output())).To(Equal("hello\n"))
}

func TestRunFailed(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/false")
	err := cmd.Run()
	g.Expect(err).NotTo(BeNil())
	failed := &FailedError{}
	g.Expect(errors.As(err, &failed)).To(BeTrue())
	g.Expect(failed.Exit).To(Equal(1))
	g.Expect(errors.Is(err, &FailedError{})).To(BeTrue())
}

func TestRunBeginError(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Options.Add("hello")
	cmd.Begin = func() (err error) {
		err = errors.New("begin failed")
		return
	}
	err := cmd.Run()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("begin failed"))
	g.Expect(cmd.Output()).To(BeEmpty())
}

func TestRunEndHook(t *testing.T) {
	g := NewGomegaWithT(t)
	called := false
	cmd := New("/bin/echo")
	cmd.End = func() {
		called = true
	}
	err := cmd.Run()
	g.Expect(err).To(BeNil())
	g.Expect(called).To(BeTrue())
}

func TestBufferWriteRead(t *testing.T) {
	g := NewGomegaWithT(t)
	buffer := &Buffer{}
	n, err := buffer.Write([]byte("abc"))
	g.Expect(err).To(BeNil())
	g.Expect(n).To(Equal(3))
	g.Expect(buffer.Bytes()).To(Equal([]byte("abc")))

	p := make([]byte, 2)
	n, err = buffer.Read(p)
	g.Expect(err).To(BeNil())
	g.Expect(n).To(Equal(2))
	g.Expect(p[:n]).To(Equal([]byte("ab")))

	n, err = buffer.Read(p)
	g.Expect(err).To(BeNil())
	g.Expect(n).To(Equal(1))
	g.Expect(p[:n]).To(Equal([]byte("c")))

	_, err = buffer.Read(p)
	g.Expect(err).To(Equal(io.EOF))
}

func TestBufferSeek(t *testing.T) {
	g := NewGomegaWithT(t)
	buffer := &Buffer{}
	_, _ = buffer.Write([]byte("abc"))
	_, _ = buffer.Read(make([]byte, 3))

	n, err := buffer.Seek(0, io.SeekStart)
	g.Expect(err).To(BeNil())
	g.Expect(n).To(Equal(int64(0)))
	p := make([]byte, 3)
	rd, err := buffer.Read(p)
	g.Expect(err).To(BeNil())
	g.Expect(p[:rd]).To(Equal([]byte("abc")))

	n, err = buffer.Seek(0, io.SeekEnd)
	g.Expect(err).To(BeNil())
	g.Expect(n).To(Equal(int64(3)))

	_, err = buffer.Seek(0, 99)
	g.Expect(err).NotTo(BeNil())

	_, err = buffer.Seek(100, io.SeekStart)
	g.Expect(err).NotTo(BeNil())
}

func TestOutput(t *testing.T) {
	g := NewGomegaWithT(t)
	cmd := New("/bin/echo")
	cmd.Writer = &Buffer{}
	_, _ = cmd.Writer.Write([]byte("data"))
	g.Expect(string(cmd.Output())).To(Equal("data"))
}
