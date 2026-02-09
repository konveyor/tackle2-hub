package simulator

import (
	core "k8s.io/api/core/v1"
)

var _ PodManager = (*StubManager)(nil)

// StubManager is a stub implementation of PodManager for testing.
type StubManager struct {
	DoCreated func(pod *core.Pod)
	DoNext    func(pod *core.Pod) (phase core.PodPhase)
	DoDeleted func(pod *core.Pod)
}

// Created adds a pod for monitoring.
func (s *StubManager) Created(pod *core.Pod) {
	if s.DoCreated != nil {
		s.DoCreated(pod)
	}
}

// Next returns the next pod phase.
func (s *StubManager) Next(pod *core.Pod) (phase core.PodPhase) {
	if s.DoNext != nil {
		phase = s.DoNext(pod)
	}
	return
}

// Deleted removes a pod from monitoring.
func (s *StubManager) Deleted(pod *core.Pod) {
	if s.DoDeleted != nil {
		s.DoDeleted(pod)
	}
}
