package simulator

import (
	"sync"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewMonitor returns a time-based pod monitor.
func NewMonitor(pending, running int) (m *TimedMonitor) {
	m = &TimedMonitor{
		podMap: make(map[types.UID]TimedPod),
	}
	m.Use(pending, running)
	return
}

// PodMonitor monitors pod progression through execution.
// Provides a method to orchestrate succeeded and failed cases.
type PodMonitor interface {
	Created(pod *core.Pod)
	Next(pod *core.Pod) (phase core.PodPhase)
	Deleted(pod *core.Pod)
}

// TimedMonitor provides time-based execution.
type TimedMonitor struct {
	mutex      sync.Mutex
	podMap     map[types.UID]TimedPod
	Thresholds struct {
		Pending time.Duration
		Running time.Duration
	}
}

// Use updates the monitor.
func (m *TimedMonitor) Use(pending, running int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Thresholds.Pending = time.Duration(pending) * time.Second
	m.Thresholds.Running = time.Duration(running) * time.Second
}

// Created adds a pod for monitoring.
func (m *TimedMonitor) Created(pod *core.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	p := make(TimedPod)
	p[pod.Status.Phase] = time.Now()
	m.podMap[pod.UID] = p
}

// Deleted deletes a pod.
func (m *TimedMonitor) Deleted(pod *core.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.podMap, pod.UID)
}

// Next returns the next pod phase.
func (m *TimedMonitor) Next(pod *core.Pod) (phase core.PodPhase) {
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
