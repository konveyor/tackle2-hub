package filter

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	filter := Filter{}
	filter.And("name").Equals([]string{"One", "Two", "Three"})
	p := filter.String()
	g.Expect("name=('One'|'Two'|'Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Equals(And{"One", "Two", "Three"})
	p = filter.String()
	g.Expect("name=('One','Two','Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Equals("Elmer")
	filter.And("age").GreaterThan(10)
	p = filter.String()
	g.Expect("name='Elmer',age>10").To(gomega.Equal(p))
}
