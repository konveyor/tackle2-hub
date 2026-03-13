package simulator

import (
	"fmt"
	"sync"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// Run attempts to run a pod and returns its resulting phase.
func (n *BaseNode) Run(pod *core.Pod) (phase core.PodPhase) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	consumed := n.Consumed
	for _, cnt := range pod.Spec.Containers {
		q := cnt.Resources.Requests.Cpu()
		if q != nil {
			consumed.cpu.Add(*q)
		}
		q = cnt.Resources.Requests.Memory()
		if q != nil {
			consumed.memory.Add(*q)
		}
	}
	phase = core.PodPending
	allocated := n.Allocated.cpu.MilliValue()
	if allocated > 0 && consumed.cpu.MilliValue() > allocated {
		n.markUnschedulable(pod, "Insufficient cpu")
		return
	}
	allocated = n.Allocated.memory.Value()
	if allocated > 0 && consumed.memory.Value() > allocated {
		n.markUnschedulable(pod, "Insufficient memory")
		return
	}
	n.Consumed = consumed
	n.clearUnschedulable(pod)
	phase = core.PodRunning
	return
}

// Terminated return pod resources.
func (n *BaseNode) Terminated(pod *core.Pod) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	consumed := n.Consumed
	for _, cnt := range pod.Spec.Containers {
		q := cnt.Resources.Requests.Cpu()
		if q != nil {
			consumed.cpu.Sub(*q)
		}
		q = cnt.Resources.Requests.Memory()
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

// markUnschedulable marks a pod as unschedulable.
func (n *BaseNode) markUnschedulable(pod *core.Pod, message string) {
	for _, cnd := range pod.Status.Conditions {
		if cnd.Type == core.PodScheduled &&
			cnd.Status == core.ConditionFalse &&
			cnd.Reason == core.PodReasonUnschedulable {
			return
		}
	}
	pod.Status.Conditions = append(
		pod.Status.Conditions,
		core.PodCondition{
			Type:               core.PodScheduled,
			Status:             core.ConditionFalse,
			Reason:             core.PodReasonUnschedulable,
			Message:            message,
			LastTransitionTime: meta_v1.Now(),
		})
}

// clearUnschedulable clears the unschedulable condition when pod is successfully scheduled.
func (n *BaseNode) clearUnschedulable(pod *core.Pod) {
	for i := range pod.Status.Conditions {
		cnd := &pod.Status.Conditions[i]
		if cnd.Type == core.PodScheduled &&
			cnd.Status == core.ConditionFalse &&
			cnd.Reason == core.PodReasonUnschedulable {
			cnd.Status = core.ConditionTrue
			cnd.Reason = ""
			cnd.Message = ""
			cnd.LastTransitionTime = meta_v1.Now()
			break
		}
	}
}
