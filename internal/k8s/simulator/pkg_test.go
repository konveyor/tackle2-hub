package simulator

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
)

// TestPodLifecycle tests pod state progression.
func TestPodLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)
	// Create simulator with fast timing for testing
	simClient := New().Use(NewManager(1, 1))
	// Create a pod
	pod := &core.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "addon",
					Image: "addon:latest",
				},
				{
					Name:  "ext0",
					Image: "extension:latest",
				},
			},
		},
	}
	// CREATE: Create the pod
	err := simClient.Create(context.TODO(), pod)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	// Immediately after creation, pod should be Pending
	retrieved := &core.Pod{}
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-pod", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(retrieved.Status.Phase).To(gomega.Equal(core.PodPending))
	// Wait for pod to transition to Running (2 seconds)
	time.Sleep(time.Second)
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-pod", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(retrieved.Status.Phase).To(gomega.Equal(core.PodRunning))
	g.Expect(retrieved.Status.ContainerStatuses).ToNot(gomega.BeEmpty())
	g.Expect(retrieved.Status.ContainerStatuses[0].Ready).To(gomega.BeTrue())
	g.Expect(retrieved.Status.ContainerStatuses[1].Ready).To(gomega.BeTrue())
	// Wait for pod to transition to Succeeded (3 more seconds)
	time.Sleep(time.Second)
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-pod", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(retrieved.Status.Phase).To(gomega.Equal(core.PodSucceeded))
	g.Expect(retrieved.Status.ContainerStatuses).ToNot(gomega.BeEmpty())
	g.Expect(retrieved.Status.ContainerStatuses[0].State.Terminated).ToNot(gomega.BeNil())
	g.Expect(retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode).To(gomega.Equal(int32(0)))
	g.Expect(retrieved.Status.ContainerStatuses[1].State.Terminated).ToNot(gomega.BeNil())
	g.Expect(retrieved.Status.ContainerStatuses[1].State.Terminated.ExitCode).To(gomega.Equal(int32(0)))
	// DELETE: Delete the pod
	err = simClient.Delete(context.TODO(), pod)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	// Verify pod is deleted
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-pod", Namespace: "default"}, retrieved)
	g.Expect(err).To(gomega.HaveOccurred())
}

// TestList tests listing resources.
func TestList(t *testing.T) {
	g := gomega.NewWithT(t)
	simClient := New()
	// Create multiple pods
	for i := 0; i < 3; i++ {
		pod := &core.Pod{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "test-pod-" + string(rune('0'+i)),
				Namespace: "default",
			},
		}
		err := simClient.Create(context.TODO(), pod)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}
	// List pods
	podList := &core.PodList{}
	err := simClient.List(context.TODO(), podList)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(podList.Items).To(gomega.HaveLen(3))
}

// TestResourceTypes tests different resource types.
func TestResourceTypes(t *testing.T) {
	g := gomega.NewWithT(t)
	simClient := New()
	// Test Secret
	secret := &core.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}
	err := simClient.Create(context.TODO(), secret)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	retrieved := &core.Secret{}
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-secret", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(retrieved.Data["key"])).To(gomega.Equal("value"))
	// Test Addon
	addon := &crd.Addon{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-addon",
			Namespace: "default",
		},
	}
	err = simClient.Create(context.TODO(), addon)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	addonList := &crd.AddonList{}
	err = simClient.List(context.TODO(), addonList)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(addonList.Items).To(gomega.HaveLen(4))
}

// TestUpdate tests resource updates.
func TestUpdate(t *testing.T) {
	g := gomega.NewWithT(t)
	simClient := New()
	// Create a secret
	secret := &core.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"key": []byte("original"),
		},
	}
	err := simClient.Create(context.TODO(), secret)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	// Update the secret
	secret.Data["key"] = []byte("updated")
	err = simClient.Update(context.TODO(), secret)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	// Verify update
	retrieved := &core.Secret{}
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "test-secret", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(string(retrieved.Data["key"])).To(gomega.Equal("updated"))
}

// TestWithFailures tests pod failure simulation.
func TestWithFailures(t *testing.T) {
	g := gomega.NewWithT(t)
	// Create simulator that always fails pods
	stub := &StubManager{
		DoNext: func(pod *core.Pod) (phase core.PodPhase) {
			return core.PodFailed
		},
	}
	simClient := New().Use(stub)
	pod := &core.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "failing-pod",
			Namespace: "default",
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "main",
					Image: "test:latest",
				},
			},
		},
	}
	err := simClient.Create(context.TODO(), pod)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	// Wait for pod to complete
	time.Sleep(2500 * time.Millisecond)
	retrieved := &core.Pod{}
	err = simClient.Get(context.TODO(), client.ObjectKey{Name: "failing-pod", Namespace: "default"}, retrieved)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(retrieved.Status.Phase).To(gomega.Equal(core.PodFailed))
	g.Expect(retrieved.Status.ContainerStatuses).ToNot(gomega.BeEmpty())
	g.Expect(retrieved.Status.ContainerStatuses[0].State.Terminated).ToNot(gomega.BeNil())
	g.Expect(retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode).ToNot(gomega.Equal(int32(0)))
}

// TestNodeResourceAllocation tests node resource allocation.
func TestNodeResourceAllocation(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "1Gi").(*BaseNode)

	g.Expect(node.Allocated.cpu.MilliValue()).To(gomega.Equal(int64(1000)))
	g.Expect(node.Allocated.memory.Value()).To(gomega.Equal(int64(1073741824))) // 1Gi
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(0)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(0)))
}

// TestNodeRunWithinCapacity tests running pods within node capacity.
func TestNodeRunWithinCapacity(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "1Gi").(*BaseNode)

	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}

	phase := node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodRunning))
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(500)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(536870912))) // 512Mi
}

// TestNodeRunExceedsCapacity tests running pods that exceed node capacity.
func TestNodeRunExceedsCapacity(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "1Gi").(*BaseNode)

	// First pod consumes most resources
	pod1 := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test1",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("800m"),
							core.ResourceMemory: *parseQuantity("768Mi"),
						},
					},
				},
			},
		},
	}
	phase := node.Run(pod1)
	g.Expect(phase).To(gomega.Equal(core.PodRunning))

	// Second pod exceeds CPU capacity
	pod2 := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test2",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("300m"),
							core.ResourceMemory: *parseQuantity("128Mi"),
						},
					},
				},
			},
		},
	}
	phase = node.Run(pod2)
	g.Expect(phase).To(gomega.Equal(core.PodPending))
	// Resources should not be consumed for failed scheduling
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(800)))
	// Pod should have unschedulable condition
	g.Expect(pod2.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod2.Status.Conditions[0].Type).To(gomega.Equal(core.PodScheduled))
	g.Expect(pod2.Status.Conditions[0].Status).To(gomega.Equal(core.ConditionFalse))
	g.Expect(pod2.Status.Conditions[0].Reason).To(gomega.Equal(core.PodReasonUnschedulable))
	g.Expect(pod2.Status.Conditions[0].Message).To(gomega.ContainSubstring("cpu"))
}

// TestNodeTerminated tests resource freeing when pods terminate.
func TestNodeTerminated(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "1Gi").(*BaseNode)

	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}

	// Run pod
	phase := node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodRunning))
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(500)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(536870912)))

	// Terminate pod
	node.Terminated(pod)
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(0)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(0)))
}

// TestNodeTerminatedNegativeProtection tests negative resource protection.
func TestNodeTerminatedNegativeProtection(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "1Gi").(*BaseNode)

	// Terminate a pod that was never run (should not go negative)
	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}

	node.Terminated(pod)
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(0)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(0)))
}

// TestNodeString tests node string representation.
func TestNodeString(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("2000m", "2Gi").(*BaseNode)

	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}
	node.Run(pod)

	str := node.String()
	g.Expect(str).To(gomega.ContainSubstring("500m"))
	g.Expect(str).To(gomega.ContainSubstring("2"))
	g.Expect(str).To(gomega.ContainSubstring("512Mi"))
}

// TestNodeUnschedulableCondition tests that unschedulable conditions are set correctly.
func TestNodeUnschedulableCondition(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("1000m", "512Mi").(*BaseNode)

	// Pod that exceeds memory
	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("1Gi"),
						},
					},
				},
			},
		},
	}

	phase := node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodPending))

	// Verify unschedulable condition
	g.Expect(pod.Status.Conditions).To(gomega.HaveLen(1))
	condition := pod.Status.Conditions[0]
	g.Expect(condition.Type).To(gomega.Equal(core.PodScheduled))
	g.Expect(condition.Status).To(gomega.Equal(core.ConditionFalse))
	g.Expect(condition.Reason).To(gomega.Equal(core.PodReasonUnschedulable))
	g.Expect(condition.Message).To(gomega.ContainSubstring("memory"))
	g.Expect(condition.LastTransitionTime.IsZero()).To(gomega.BeFalse())
}

// TestNodeClearUnschedulable tests that unschedulable condition is cleared when pod is scheduled.
func TestNodeClearUnschedulable(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("400m", "256Mi").(*BaseNode) // Not enough for the pod

	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "test",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("500m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}

	// First attempt: Pod won't fit (exceeds capacity)
	phase := node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodPending))
	g.Expect(pod.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod.Status.Conditions[0].Reason).To(gomega.Equal(core.PodReasonUnschedulable))
	g.Expect(pod.Status.Conditions[0].Status).To(gomega.Equal(core.ConditionFalse))

	// Reconfigure node with enough resources
	node = node.With("2000m", "2Gi").(*BaseNode)

	// Second attempt: Pod should fit now
	phase = node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodRunning))

	// Verify unschedulable condition was cleared (changed to True)
	g.Expect(pod.Status.Conditions).To(gomega.HaveLen(1))
	g.Expect(pod.Status.Conditions[0].Type).To(gomega.Equal(core.PodScheduled))
	g.Expect(pod.Status.Conditions[0].Status).To(gomega.Equal(core.ConditionTrue))
	g.Expect(pod.Status.Conditions[0].Reason).To(gomega.Equal(""))
	g.Expect(pod.Status.Conditions[0].Message).To(gomega.Equal(""))
}

// TestNodeMultipleContainers tests resource tracking with multiple containers.
func TestNodeMultipleContainers(t *testing.T) {
	g := gomega.NewWithT(t)
	node := &BaseNode{}
	node = node.With("2000m", "2Gi").(*BaseNode)

	pod := &core.Pod{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "container1",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("300m"),
							core.ResourceMemory: *parseQuantity("256Mi"),
						},
					},
				},
				{
					Name:  "container2",
					Image: "test:latest",
					Resources: core.ResourceRequirements{
						Requests: core.ResourceList{
							core.ResourceCPU:    *parseQuantity("400m"),
							core.ResourceMemory: *parseQuantity("512Mi"),
						},
					},
				},
			},
		},
	}

	phase := node.Run(pod)
	g.Expect(phase).To(gomega.Equal(core.PodRunning))
	// Total: 300m + 400m = 700m CPU, 256Mi + 512Mi = 768Mi memory
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(700)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(805306368))) // 768Mi

	node.Terminated(pod)
	g.Expect(node.Consumed.cpu.MilliValue()).To(gomega.Equal(int64(0)))
	g.Expect(node.Consumed.memory.Value()).To(gomega.Equal(int64(0)))
}

// parseQuantity is a helper to parse resource quantities.
func parseQuantity(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}
