package filter

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	filter := Filter{}
	list := Any{"One", "Two", "Three"}
	filter.And("name").Equals(list)
	p := filter.String()
	g.Expect("name=('One'|'Two'|'Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Equals(And{"One", "Two", "Three"})
	p = filter.String()
	g.Expect("name=('One','Two','Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Equals("Elmer")
	filter.And("age").GreaterThan(10)
	filter.And("height").LessThan(44)
	p = filter.String()
	g.Expect("name='Elmer',age>10,height<44").To(gomega.Equal(p))
}
