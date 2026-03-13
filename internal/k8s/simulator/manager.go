package simulator

import (
	"fmt"
	"sync"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

// NewManager returns a time-based pod manager.
func NewManager(pending, running int) (m *TimedManager) {
	m = &TimedManager{
		podMap: make(map[types.UID]TimedPod),
		Node:   &BaseNode{},
	}
	m.Use(pending, running)
	return
}

// Resources are allocated resources.
type Resources struct {
	cpu    resource.Quantity
	memory resource.Quantity
}

// Node simulates a k8s node.
type Node interface {
	// With configures node resources and returns a new node instance.
	With(cpu, memory string) Node
	// Run attempts to run a pod and returns its resulting phase.
	Run(pod *core.Pod) (phase core.PodPhase)
	// Terminated releases resources consumed by a terminated pod.
	Terminated(pod *core.Pod)
	// String returns the node's resource usage.
	String() string
}

// BaseNode simulates a k8s node with resource tracking.
type BaseNode struct {
	mutex     sync.Mutex
	Allocated Resources
	Consumed  Resources
}

// With resources.
func (n *BaseNode) With(cpu, memory string) (n2 Node) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	b := &BaseNode{}
	b.Allocated.cpu = resource.MustParse(cpu)
	b.Allocated.memory = resource.MustParse(memory)
	n2 = b
	return
}

// Run resources.
func (n *BaseNode) Run(pod *core.Pod) (phase core.PodPhase) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	consumed := n.Consumed
	for _, cnt := range pod.Spec.Containers {
		q := cnt.Resources.Limits.Cpu()
		if q != nil {
			consumed.cpu.Add(*q)
		}
		q = cnt.Resources.Limits.Memory()
		if q != nil {
			consumed.memory.Add(*q)
		}
	}
	phase = core.PodPending
	allocated := n.Allocated.cpu.Value()
	if allocated > 0 && consumed.cpu.Value() > allocated {
		return
	}
	allocated = n.Allocated.memory.Value()
	if allocated > 0 && consumed.memory.Value() > allocated {
		return
	}
	n.Consumed = consumed
	phase = core.PodRunning
	return
}

// Terminated return pod resources.
func (n *BaseNode) Terminated(pod *core.Pod) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	consumed := n.Consumed
	for _, cnt := range pod.Spec.Containers {
		q := cnt.Resources.Limits.Cpu()
		if q != nil {
			consumed.cpu.Sub(*q)
		}
		q = cnt.Resources.Limits.Memory()
		if q != nil {
			consumed.memory.Sub(*q)
		}
	}
	if consumed.cpu.Sign() < 0 {
		consumed.cpu.Set(0)
	}
	if consumed.memory.Sign() < 0 {
		consumed.memory.Set(0)
	}
	n.Consumed = consumed
}

// String returns the node's resource usage.
func (n *BaseNode) String() (s string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	s = fmt.Sprintf(
		"(Node) cpu: %s/%s memory: %s/%s",
		n.Consumed.cpu.String(),
		n.Allocated.cpu.String(),
		n.Consumed.memory.String(),
		n.Allocated.memory.String())
	return
}

// PodManager managers pod progression through execution.
// Provides a method to orchestrate succeeded and failed cases.
type PodManager interface {
	Created(pod *core.Pod)
	Next(pod *core.Pod) (phase core.PodPhase)
	Deleted(pod *core.Pod)
}

// TimedManager provides time-based execution.
type TimedManager struct {
	mutex      sync.Mutex
	podMap     map[types.UID]TimedPod
	Node       Node
	Thresholds struct {
		Pending time.Duration
		Running time.Duration
	}
}

// Use updates the manager.
func (m *TimedManager) Use(pending, running int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Thresholds.Pending = time.Duration(pending) * time.Second
	m.Thresholds.Running = time.Duration(running) * time.Second
}

// Created adds a pod for managing.
func (m *TimedManager) Created(pod *core.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	p := make(TimedPod)
	p[pod.Status.Phase] = time.Now()
	m.podMap[pod.UID] = p
}

// Deleted deletes a pod.
func (m *TimedManager) Deleted(pod *core.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.podMap, pod.UID)
}

// Next returns the next pod phase.
func (m *TimedManager) Next(pod *core.Pod) (phase core.PodPhase) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	phase = pod.Status.Phase
	p, found := m.podMap[pod.UID]
	if !found {
		phase = core.PodFailed
		return
	}
	switch phase {
	case core.PodPending:
		if p.Exceeded(pod, m.Thresholds.Pending) {
			phase = m.Node.Run(pod)
			p.Mark(phase)
		}
	case core.PodRunning:
		if p.Exceeded(pod, m.Thresholds.Running) {
			m.Node.Terminated(pod)
			phase = core.PodSucceeded
			p.Mark(phase)
		}
	case core.PodSucceeded:
		//
	case core.PodFailed:
		//
	default:
		phase = core.PodFailed
		p.Mark(phase)
	}

	return
}

// TimedPod timed pod.
type TimedPod map[core.PodPhase]time.Time

func (m TimedPod) Mark(phase core.PodPhase) {
	m[phase] = time.Now()
}

// Exceeded returns true when the threshold has been exceeded.
func (m TimedPod) Exceeded(pod *core.Pod, threshold time.Duration) (b bool) {
	d, found := m[pod.Status.Phase]
	if !found {
		m.Mark(pod.Status.Phase)
		return
	}
	b = time.Since(d) > threshold
	return
}
