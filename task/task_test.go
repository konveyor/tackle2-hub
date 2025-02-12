package task

import (
	"testing"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/onsi/gomega"
)

func TestPriorityEscalate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	appId := uint(1)
	appOther := uint(2)

	kinds := make(map[string]*crd.Task)
	ready := []*Task{}

	a := crd.Task{}
	a.Name = "a"
	kinds[a.Name] = &a

	b := crd.Task{}
	b.Name = "b"
	b.Spec.Dependencies = []string{"a"}
	kinds[b.Name] = &b

	c := crd.Task{}
	c.Name = "c"
	c.Spec.Dependencies = []string{"b"}
	kinds[c.Name] = &c

	task := &Task{&model.Task{}}
	task.ID = 1
	task.Kind = "c"
	task.State = Ready
	task.Priority = 10
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 2
	task.Kind = "b"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 3
	task.Kind = "a"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 4
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 5
	task.Kind = "b"
	task.State = Ready
	task.ApplicationID = &appOther
	ready = append(ready, task)

	pE := Priority{
		cluster: &Cluster{
			tasks: kinds,
		}}

	escalated := pE.Escalate(ready)
	g.Expect(len(escalated)).To(gomega.Equal(2))

	escalated = pE.Escalate(nil)
	g.Expect(len(escalated)).To(gomega.Equal(0))
}

func TestPriorityGraph(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	appId := uint(1)

	kinds := make(map[string]*crd.Task)
	ready := []*Task{}

	a := crd.Task{}
	a.Name = "a"
	kinds[a.Name] = &a

	b := crd.Task{}
	b.Name = "b"
	b.Spec.Dependencies = []string{"a"}
	kinds[b.Name] = &b

	c := crd.Task{}
	c.Name = "c"
	c.Spec.Dependencies = []string{"b"}
	kinds[c.Name] = &c

	task := &Task{&model.Task{}}
	task.ID = 1
	task.Kind = "c"
	task.State = Ready
	task.Priority = 10
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 2
	task.Kind = "b"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 3
	task.Kind = "a"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = &Task{&model.Task{}}
	task.ID = 4
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	pE := Priority{
		cluster: &Cluster{
			tasks: kinds,
		}}
	deps := pE.graph(ready[0], ready)
	g.Expect(len(deps)).To(gomega.Equal(2))
}

func TestAddonRegex(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	m := Task{}
	addonA := &crd.Addon{}
	addonA.Name = "A"
	addonB := &crd.Addon{}
	addonB.Name = "B"
	// direct.
	ext := &crd.Extension{}
	ext.Name = "Test"
	ext.Spec.Addon = "A"
	matched, err := m.matchAddon(ext, addonA)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(matched).To(gomega.BeTrue())
	matched, err = m.matchAddon(ext, addonB)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(matched).To(gomega.BeFalse())
	// regex.
	ext.Spec.Addon = "^(A|B)$"
	matched, err = m.matchAddon(ext, addonA)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(matched).To(gomega.BeTrue())
	matched, err = m.matchAddon(ext, addonB)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(matched).To(gomega.BeTrue())
	// regex not valid.
	ext.Spec.Addon = "(]$"
	matched, err = m.matchAddon(ext, addonA)
	g.Expect(err).ToNot(gomega.BeNil())
}
