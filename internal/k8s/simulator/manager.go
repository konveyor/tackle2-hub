package simulator

import (
	"sync"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewManager returns a time-based pod manager.
func NewManager(pending, running int) (m *TimedManager) {
	m = &TimedManager{
		podMap: make(map[types.UID]TimedPod),
	}
	m.Use(pending, running)
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
			phase = core.PodRunning
			p.Mark(phase)
		}
	case core.PodRunning:
		if p.Exceeded(pod, m.Thresholds.Running) {
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
