package tasking

import (
	"sync"

	core "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestPodManager is a configurable pod manager for testing various failure scenarios.
// Configure behavior by setting the appropriate fields:
//   - failFirstPod: true - first pod fails immediately
//   - unschedulable: true - pods become unschedulable
//   - killCount: N - kill pod N times with exit code 137 before success
//   - imageError: "reason" - set image pull error (ErrImagePull, ImagePullBackOff, InvalidImageName)
type TestPodManager struct {
	mutex sync.Mutex
	// Failure modes
	failFirstPod  bool   // Make first pod fail immediately
	unschedulable bool   // Make pods unschedulable
	killCount     int    // Number of times to kill pod with exit 137 before success
	imageError    string // Image error reason (ErrImagePull, ImagePullBackOff, InvalidImageName)
	// Internal state
	hasFailed         bool
	failedPod         string
	unschedulablePods map[string]bool
	currentKills      int
}

// Created is called when a pod is created.
func (m *TestPodManager) Created(pod *core.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.unschedulable {
		if m.unschedulablePods == nil {
			m.unschedulablePods = make(map[string]bool)
		}
		m.unschedulablePods[pod.Name] = true
		pod.Status.Conditions = []core.PodCondition{
			{
				Type:   core.PodScheduled,
				Status: core.ConditionFalse,
				Reason: core.PodReasonUnschedulable,
			},
		}
	}
}

// Next returns the next phase for the pod based on configured behavior.
func (m *TestPodManager) Next(pod *core.Pod) (phase core.PodPhase) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Handle unschedulable pods
	if m.unschedulable && m.unschedulablePods[pod.Name] {
		phase = core.PodPending
		return
	}

	// Handle image pull errors
	if m.imageError != "" && pod.Status.Phase == core.PodPending {
		statuses := make([]core.ContainerStatus, len(pod.Spec.Containers))
		for i, container := range pod.Spec.Containers {
			statuses[i] = core.ContainerStatus{
				Name:  container.Name,
				Ready: false,
				State: core.ContainerState{
					Waiting: &core.ContainerStateWaiting{
						Reason:  m.imageError,
						Message: "Failed to pull image: " + m.imageError,
					},
				},
			}
		}
		pod.Status.ContainerStatuses = statuses
		phase = core.PodPending
		return
	}

	// Determine if this pod should fail (failFirstPod mode)
	shouldFail := false
	if m.failFirstPod {
		if !m.hasFailed {
			m.hasFailed = true
			m.failedPod = pod.Name
			shouldFail = true
		} else if pod.Name == m.failedPod {
			shouldFail = true
		}
	}

	// Transition based on current phase
	switch pod.Status.Phase {
	case core.PodPending:
		if shouldFail {
			phase = core.PodFailed
			return
		}
		phase = core.PodRunning
		return
	case core.PodRunning:
		// Handle container kill scenario
		if m.killCount > 0 && m.currentKills < m.killCount {
			m.currentKills++
			statuses := make([]core.ContainerStatus, len(pod.Spec.Containers))
			for i, container := range pod.Spec.Containers {
				statuses[i] = core.ContainerStatus{
					Name:  container.Name,
					Ready: false,
					State: core.ContainerState{
						Terminated: &core.ContainerStateTerminated{
							ExitCode:   137, // SIGKILL
							Reason:     "Killed",
							Message:    "Container was killed",
							FinishedAt: meta_v1.Now(),
						},
					},
				}
			}
			pod.Status.ContainerStatuses = statuses
			phase = core.PodFailed
			return
		}
		if shouldFail {
			phase = core.PodFailed
			return
		}
		phase = core.PodSucceeded
		return
	default:
		phase = pod.Status.Phase
		return
	}
}

// Deleted is called when a pod is deleted.
func (m *TestPodManager) Deleted(pod *core.Pod) {
	// No-op
}
