package scm

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestSvnURL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	remote := Remote{
		URL:    "http://svn.corp/project",
		Branch: "trunk",
		Path:   "eng/product/thing/app_1",
	}
	// Load.
	u := SvnURL{}
	err := u.With(remote)
	g.Expect(err).To(gomega.BeNil())
	// String()
	s := u.String()
	expectStr := "http://svn.corp/project/trunk/eng/product/thing/app_1"
	g.Expect(expectStr).To(gomega.Equal(s))
}
