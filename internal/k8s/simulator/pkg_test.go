package simulator

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
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
	simClient := New().Use(&TestManager{})
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

type TestManager struct {
	PodManager
}

func (m *TestManager) Created(pod *core.Pod) {}

func (m *TestManager) Deleted(pod *core.Pod) {}

func (m *TestManager) Next(pod *core.Pod) (next core.PodPhase) {
	return core.PodFailed
}
