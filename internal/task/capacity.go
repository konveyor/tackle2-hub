package task

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	core "k8s.io/api/core/v1"
)

// CapacityMonitor determines cluster capacity.
// Maintains a running estimation of the cluster scheduling capacity.
type CapacityMonitor struct {
	mutex       sync.Mutex
	background  atomic.Bool
	capacity    int
	scheduled   int
	unscheduled int
	growthRate  float64
}

// Reset the statistics.
func (m *CapacityMonitor) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.capacity = 1
	m.scheduled = 0
	m.unscheduled = 0
	m.growthRate = 1.05
}

// Run the monitor.
func (m *CapacityMonitor) Run(ctx context.Context, cluster *Cluster) {
	if m.Background() {
		return
	}
	m.Reset()
	go func() {
		Log.Info("CapacityMonitor started.")
		m.background.Store(true)
		defer func() {
			Log.Info("CapacityMonitor stopped.")
			m.background.Store(false)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.pause()
				err := m.Reconcile(cluster)
				if err != nil {
					Log.Error(err, "")
				}
			}
		}
	}()
}

// Background returns true when running in a goroutine.
func (m *CapacityMonitor) Background() (started bool) {
	return m.background.Load()
}

// Reconcile capacity.
// Turns the 'crank' once.
// Intended to be called directly from Run() and test harnesses.
func (m *CapacityMonitor) Reconcile(cluster *Cluster) (err error) {
	err = cluster.Refresh()
	if err != nil {
		return
	}
	m.Adjust(cluster)
	return
}

// Adjust updates estimated cluster capacity.
func (m *CapacityMonitor) Adjust(cluster *Cluster) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	scheduled := 0
	unscheduled := 0
	digest := m.digest()
	for _, pod := range cluster.TaskPods() {
		switch pod.Status.Phase {
		case core.PodPending:
			match := false
			for _, p := range pod.Status.Conditions {
				if p.Type == core.PodScheduled {
					if p.Reason == core.PodReasonUnschedulable {
						match = true
						break
					}
				}
			}
			if match {
				unscheduled++
			} else {
				scheduled++
			}
		case core.PodRunning:
			scheduled++
		}
	}
	if scheduled == 0 && m.capacity > 0 {
		m.scheduled = scheduled
		m.unscheduled = unscheduled
		return
	}
	if unscheduled == 0 {
		next := float64(scheduled)
		next *= m.growthRate
		next = max(next, 1.0)
		nextInt := int(next + 0.9999)
		m.capacity = max(m.capacity, nextInt)
	} else {
		if unscheduled > m.unscheduled {
			m.capacity = scheduled - (unscheduled - m.unscheduled)
		}
	}
	m.scheduled = scheduled
	m.unscheduled = unscheduled
	m.capacity = max(m.capacity, 0)
	if m.digest() != digest {
		Log.Info("Capacity adjusted: " + m.string())
	}
}

// Current returns current capacity.
func (m *CapacityMonitor) Current() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.capacity
}

// Exceeded returns true when unscheduled pods are detected.
func (m *CapacityMonitor) Exceeded() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.unscheduled > 0
}

// String returns a string representation.
func (m *CapacityMonitor) String() (s string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s = m.string()
	return
}

// String returns a string representation.
func (m *CapacityMonitor) string() (s string) {
	s = fmt.Sprintf(
		"[pods] capacity:%d,scheduled:%d",
		m.capacity,
		m.scheduled)
	if m.unscheduled > 0 {
		s += fmt.Sprintf(",UNSCHEDULED:%d", m.unscheduled)
	}
	return
}

// digest returns a digest.
func (m *CapacityMonitor) digest() (d string) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strconv.Itoa(m.capacity)))
	_, _ = h.Write([]byte(strconv.Itoa(m.scheduled)))
	_, _ = h.Write([]byte(strconv.Itoa(m.unscheduled)))
	n := h.Sum64()
	d = fmt.Sprintf("%x", n)
	return
}

// Pause.
func (m *CapacityMonitor) pause() {
	time.Sleep(Settings.Frequency.Task)
}
