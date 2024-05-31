package task

import (
	"testing"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/onsi/gomega"
)

func TestPriorityEscalation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	appId := uint(1)

	kinds := make(map[string]*crd.Task)
	ready := []*Task{}

	k := crd.Task{}
	k.Name = "a"
	kinds[k.Name] = &k

	k = crd.Task{}
	k.Name = "b"
	k.Spec.Dependencies = []string{"a"}
	kinds[k.Name] = &k

	k = crd.Task{}
	k.Name = "c"
	k.Spec.Dependencies = []string{"b"}
	kinds[k.Name] = &k

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
		cluster: Cluster{
			tasks: kinds,
		}}

	escalated := pE.Escalate(ready)
	g.Expect(len(escalated)).To(gomega.Equal(2))

	escalated = pE.Escalate(nil)
	g.Expect(len(escalated)).To(gomega.Equal(0))
}
