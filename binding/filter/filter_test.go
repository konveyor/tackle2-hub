package filter

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	filter := Filter{}
	filter.Equals("name", []string{"One", "Two", "Three"})
	p := filter.String()
	g.Expect("name=('One'|'Two'|'Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.Equals("name", And{"One", "Two", "Three"})
	p = filter.String()
	g.Expect("name=('One','Two','Three')").To(gomega.Equal(p))
}
