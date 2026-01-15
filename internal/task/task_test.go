package task

import (
	"bytes"
	"io"
	"strconv"
	"testing"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/model"
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

	task := NewTask(&model.Task{})
	task.ID = 1
	task.Kind = "c"
	task.State = Ready
	task.Priority = 10
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 2
	task.Kind = "b"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 3
	task.Kind = "a"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 4
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
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
	platformId := uint(2)

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

	//
	// subject: application
	task := NewTask(&model.Task{})
	task.ID = 1
	task.Kind = "c"
	task.State = Ready
	task.Priority = 10
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 2
	task.Kind = "b"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 3
	task.Kind = "a"
	task.State = Ready
	task.ApplicationID = &appId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
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
	//
	// subject: platform
	task = NewTask(&model.Task{})
	task.ID = 1
	task.Kind = "c"
	task.State = Ready
	task.Priority = 10
	task.PlatformID = &platformId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 2
	task.Kind = "b"
	task.State = Ready
	task.PlatformID = &platformId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 3
	task.Kind = "a"
	task.State = Ready
	task.PlatformID = &platformId
	ready = append(ready, task)

	task = NewTask(&model.Task{})
	task.ID = 4
	task.State = Ready
	task.PlatformID = &platformId
	ready = append(ready, task)

	pE = Priority{
		cluster: &Cluster{
			tasks: kinds,
		}}
	deps = pE.graph(ready[0], ready)
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

func TestLogCollectorCopy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// no skipped bytes.
	collector := LogCollector{}
	content := "ABCDEFGHIJ"
	reader := io.NopCloser(bytes.NewBufferString(content))
	writer := bytes.NewBufferString("")
	err := collector.copy(reader, writer)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.Equal(writer.String()))

	// number of skipped bytes smaller than buffer.
	existing := "ABC"
	collector = LogCollector{
		nSkip: int64(len(existing)),
	}
	content = "ABCDEFGHIJ"
	reader = io.NopCloser(bytes.NewBufferString(content))
	writer = bytes.NewBufferString(existing)
	err = collector.copy(reader, writer)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.Equal(writer.String()))

	// number of skipped bytes larger than buffer.
	existing = "ABCD"
	collector = LogCollector{
		nBuf:  3,
		nSkip: int64(len(existing)),
	}
	content = "ABCDEFGHIJ"
	reader = io.NopCloser(bytes.NewBufferString(content))
	writer = bytes.NewBufferString(existing)
	err = collector.copy(reader, writer)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.Equal(writer.String()))

	// number of skipped bytes equals buffer.
	existing = "ABCD"
	collector = LogCollector{
		nBuf:  len(existing),
		nSkip: int64(len(existing)),
	}
	content = "ABCDEFGHIJ"
	reader = io.NopCloser(bytes.NewBufferString(content))
	writer = bytes.NewBufferString(existing)
	err = collector.copy(reader, writer)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(content).To(gomega.Equal(writer.String()))
}

func TestRuleIsolated(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rule := RuleIsolated{}
	tasks := []*Task{
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
	}
	tasks[0].ID = 1
	tasks[0].Policy = model.TaskPolicy{Isolated: true}
	tasks[1].ID = 2

	domain := NewDomain(tasks)

	matched, _ := rule.Match(tasks[0], domain)
	g.Expect(matched).To(gomega.BeFalse())

	matched, _ = rule.Match(tasks[1], domain)
	g.Expect(matched).To(gomega.BeTrue())
}

func TestRuleUnique(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rule := RuleUnique{
		matched: make(map[uint]uint),
	}
	tasks := []*Task{
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
	}
	one := uint(1)
	two := uint(2)

	tasks[0].ID = 1
	tasks[0].Kind = "A"
	tasks[0].ApplicationID = &one

	tasks[1].ID = 2
	tasks[1].Kind = "A"
	tasks[1].ApplicationID = &one

	tasks[2].ID = 3
	tasks[2].Kind = "B"
	tasks[2].ApplicationID = &two

	tasks[3].ID = 4
	tasks[3].Kind = "A"
	tasks[3].PlatformID = &one

	tasks[4].ID = 5
	tasks[4].Kind = "A"
	tasks[4].PlatformID = &one

	tasks[5].ID = 6
	tasks[5].Kind = "B"
	tasks[5].PlatformID = &two

	domain := NewDomain(tasks)

	// 0 matches 1 on kind
	matched, _ := rule.Match(tasks[0], domain)
	g.Expect(matched).To(gomega.BeTrue())
	g.Expect(rule.matched[tasks[0].ID]).To(gomega.Equal(tasks[1].ID))
	matched, _ = rule.Match(tasks[1], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 2 not matched.  different kind.
	matched, _ = rule.Match(tasks[2], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 0 and 1 no longer matched on addon.
	tasks[0].Kind = ""
	tasks[0].Addon = "A"
	tasks[1].Addon = "B"
	matched, _ = rule.Match(tasks[0], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 3 and 4 match on kind
	matched, _ = rule.Match(tasks[3], domain)
	g.Expect(matched).To(gomega.BeTrue())
	g.Expect(rule.matched[tasks[3].ID]).To(gomega.Equal(tasks[4].ID))
	matched, _ = rule.Match(tasks[4], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 4 not matched.  different kind.
	matched, _ = rule.Match(tasks[5], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 3 and 5 no longer matched on addon.
	tasks[3].Kind = ""
	tasks[3].Addon = "A"
	tasks[4].Addon = "B"
	matched, _ = rule.Match(tasks[3], domain)
	g.Expect(matched).To(gomega.BeFalse())
}

func TestRuleDeps(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rule := RuleDeps{}
	rule.cluster = &Cluster{}
	rule.cluster.tasks = make(map[string]*crd.Task)
	rule.cluster.tasks["B"] = &crd.Task{
		Spec: crd.TaskSpec{
			Dependencies: []string{"A"},
		}}

	tasks := []*Task{
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
		NewTask(&model.Task{}),
	}
	one := uint(1)
	two := uint(2)

	tasks[0].ID = 1
	tasks[0].Kind = "A"
	tasks[0].ApplicationID = &one

	tasks[1].ID = 2
	tasks[1].Kind = "B"
	tasks[1].ApplicationID = &one

	tasks[2].ID = 3
	tasks[2].Kind = "B"
	tasks[2].ApplicationID = &two

	tasks[3].ID = 4
	tasks[3].Kind = "B"
	tasks[3].PlatformID = &one

	tasks[4].ID = 5
	tasks[4].Kind = "B"
	tasks[4].PlatformID = &two

	tasks[5].ID = 6
	tasks[5].Kind = "C"
	tasks[5].PlatformID = &one

	domain := NewDomain(tasks)

	// no deps
	matched, _ := rule.Match(tasks[0], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 1(B) depends on 0(A)
	matched, _ = rule.Match(tasks[1], domain)
	g.Expect(matched).To(gomega.BeTrue())

	// 2(B) depends on 0(A) but different subject
	matched, _ = rule.Match(tasks[2], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 3(B) depends on 0(A) but different subject
	matched, _ = rule.Match(tasks[3], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 4(B) not depends on 0(A) and different subject
	matched, _ = rule.Match(tasks[4], domain)
	g.Expect(matched).To(gomega.BeFalse())

	// 5(C) not depends on 0(A) and different subject
	matched, _ = rule.Match(tasks[5], domain)
	g.Expect(matched).To(gomega.BeFalse())
}

func TestPredRegex(t *testing.T) {
	// Test PredRegex pattern
	g := gomega.NewGomegaWithT(t)

	matches := PredRegex.FindAllStringSubmatch("tag:Language=Java", -1)
	g.Expect(len(matches)).To(gomega.Equal(1))
	g.Expect(matches[0][1]).To(gomega.Equal("tag"))
	g.Expect(matches[0][2]).To(gomega.Equal("Language=Java"))

	matches = PredRegex.FindAllStringSubmatch("tag:Language=C#", -1)
	g.Expect(len(matches)).To(gomega.Equal(1))
	g.Expect(matches[0][1]).To(gomega.Equal("tag"))
	g.Expect(matches[0][2]).To(gomega.Equal("Language=C#"))

	// Negative tests - should NOT match
	matches = PredRegex.FindAllStringSubmatch("tag:", -1)
	g.Expect(len(matches)).To(gomega.Equal(0))

	matches = PredRegex.FindAllStringSubmatch("tag: ", -1)
	g.Expect(len(matches)).To(gomega.Equal(0))

	matches = PredRegex.FindAllStringSubmatch("tagLanguage=Java", -1)
	g.Expect(len(matches)).To(gomega.Equal(0))

	matches = PredRegex.FindAllStringSubmatch(":Language=Java", -1)
	g.Expect(len(matches)).To(gomega.Equal(0))
}

func TestSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	selector := Selector{}
	selector.predicate = map[string]Predicate{
		"T": _TestPredicate{},
	}
	m, err := selector.Match("T:true && T:true")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m).To(gomega.BeTrue())
	m, err = selector.Match("T:true && T:false")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m).To(gomega.BeFalse())
	m, err = selector.Match("T:true||T:true")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m).To(gomega.BeTrue())
	m, err = selector.Match("T:true || T:false")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m).To(gomega.BeTrue())
	m, err = selector.Match("T:false || T:false")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(m).To(gomega.BeFalse())
}

type _TestPredicate struct {
}

func (p _TestPredicate) Match(ref string) (matched bool, err error) {
	matched, _ = strconv.ParseBool(ref)
	return
}
