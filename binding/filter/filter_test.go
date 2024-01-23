package filter

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestFilter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	filter := Filter{}
	list := Any{"One", "Two", "Three"}
	filter.And("name").Eq(list)
	p := filter.String()
	g.Expect("name=('One'|'Two'|'Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Eq(All{"One", "Two", "Three"})
	p = filter.String()
	g.Expect("name=('One','Two','Three')").To(gomega.Equal(p))

	filter = Filter{}
	filter.And("name").Eq("Elmer")
	filter.And("age").Gt(10)
	filter.And("height").Lt(44)
	filter.And("weight").LtEq(150)
	filter.And("hair").NotEq("blond")
	p = filter.String()
	g.Expect("name='Elmer',age>10,height<44,weight<=150,hair!='blond'").To(gomega.Equal(p))
}
