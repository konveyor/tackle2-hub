package simulator

import (
	"context"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
)

// TestPodLifecycle tests pod state progression.
func TestPodLifecycle(t *testing.T) {
	// Create simulator with fast timing for testing
	client := New().WithTiming(2*time.Second, 3*time.Second)

	// Create a pod
	pod := &core.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-pod",
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

	// CREATE: Create the pod
	err := client.Create(context.TODO(), pod)
	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// Immediately after creation, pod should be Pending
	retrieved := &core.Pod{}
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-pod"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}
	if retrieved.Status.Phase != core.PodPending {
		t.Errorf("Expected pod phase to be Pending, got %s", retrieved.Status.Phase)
	}

	// Wait for pod to transition to Running (2 seconds)
	time.Sleep(2500 * time.Millisecond)
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-pod"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}
	if retrieved.Status.Phase != core.PodRunning {
		t.Errorf("Expected pod phase to be Running, got %s", retrieved.Status.Phase)
	}
	if len(retrieved.Status.ContainerStatuses) == 0 || !retrieved.Status.ContainerStatuses[0].Ready {
		t.Error("Expected container to be ready")
	}

	// Wait for pod to transition to Succeeded (3 more seconds)
	time.Sleep(3500 * time.Millisecond)
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-pod"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}
	if retrieved.Status.Phase != core.PodSucceeded {
		t.Errorf("Expected pod phase to be Succeeded, got %s", retrieved.Status.Phase)
	}
	if len(retrieved.Status.ContainerStatuses) == 0 {
		t.Error("Expected container statuses")
	} else if retrieved.Status.ContainerStatuses[0].State.Terminated == nil {
		t.Error("Expected container to be terminated")
	} else if retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode)
	}

	// DELETE: Delete the pod
	err = client.Delete(context.TODO(), pod)
	if err != nil {
		t.Fatalf("Failed to delete pod: %v", err)
	}

	// Verify pod is deleted
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-pod"}, retrieved)
	if err == nil {
		t.Error("Expected error when getting deleted pod")
	}
}

// TestList tests listing resources.
func TestList(t *testing.T) {
	client := New()

	// Create multiple pods
	for i := 0; i < 3; i++ {
		pod := &core.Pod{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "test-pod-" + string(rune('0'+i)),
				Namespace: "default",
			},
		}
		err := client.Create(context.TODO(), pod)
		if err != nil {
			t.Fatalf("Failed to create pod %d: %v", i, err)
		}
	}

	// List pods
	podList := &core.PodList{}
	err := client.List(context.TODO(), podList)
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	if len(podList.Items) != 3 {
		t.Errorf("Expected 3 pods, got %d", len(podList.Items))
	}
}

// TestResourceTypes tests different resource types.
func TestResourceTypes(t *testing.T) {
	client := New()

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
	err := client.Create(context.TODO(), secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	retrieved := &core.Secret{}
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-secret"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}
	if string(retrieved.Data["key"]) != "value" {
		t.Error("Secret data mismatch")
	}

	// Test Addon
	addon := &crd.Addon{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-addon",
			Namespace: "default",
		},
	}
	err = client.Create(context.TODO(), addon)
	if err != nil {
		t.Fatalf("Failed to create addon: %v", err)
	}

	addonList := &crd.AddonList{}
	err = client.List(context.TODO(), addonList)
	if err != nil {
		t.Fatalf("Failed to list addons: %v", err)
	}
	if len(addonList.Items) != 1 {
		t.Errorf("Expected 1 addon, got %d", len(addonList.Items))
	}
}

// TestUpdate tests resource updates.
func TestUpdate(t *testing.T) {
	client := New()

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
	err := client.Create(context.TODO(), secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
	}

	// Update the secret
	secret.Data["key"] = []byte("updated")
	err = client.Update(context.TODO(), secret)
	if err != nil {
		t.Fatalf("Failed to update secret: %v", err)
	}

	// Verify update
	retrieved := &core.Secret{}
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "test-secret"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}
	if string(retrieved.Data["key"]) != "updated" {
		t.Errorf("Expected 'updated', got '%s'", string(retrieved.Data["key"]))
	}
}

// TestWithFailures tests pod failure simulation.
func TestWithFailures(t *testing.T) {
	// Create simulator that always fails pods
	client := New().WithTiming(1*time.Second, 1*time.Second).WithFailureProbability(1.0)
	pod := &core.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "failing-pod",
			Namespace: "default",
		},
	}

	err := client.Create(context.TODO(), pod)
	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// Wait for pod to complete
	time.Sleep(2500 * time.Millisecond)

	retrieved := &core.Pod{}
	err = client.Get(context.TODO(), k8s.ObjectKey{Name: "failing-pod"}, retrieved)
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}

	if retrieved.Status.Phase != core.PodFailed {
		t.Errorf("Expected pod to fail, got phase %s", retrieved.Status.Phase)
	}

	if len(retrieved.Status.ContainerStatuses) > 0 {
		if retrieved.Status.ContainerStatuses[0].State.Terminated == nil {
			t.Error("Expected terminated container")
		} else if retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode == 0 {
			t.Error("Expected non-zero exit code for failed pod")
		}
	}
}
