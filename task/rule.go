package task

import (
	"fmt"
	"time"
)

// Rule defines postpone rules.
type Rule interface {
	Match(ready, other *Task) (matched bool, reason string)
}

// RuleUnique running tasks must be unique by:
//   - application
//   - addon.
type RuleUnique struct {
	matched map[uint]uint
}

// Match determines the match.
func (r *RuleUnique) Match(ready, other *Task) (matched bool, reason string) {
	if ready.ApplicationID == nil || other.ApplicationID == nil {
		return
	}
	if *ready.ApplicationID != *other.ApplicationID {
		return
	}
	if ready.Addon != other.Addon {
		return
	}
	if _, found := r.matched[other.ID]; found {
		return
	}
	matched = true
	r.matched[ready.ID] = other.ID
	reason = fmt.Sprintf(
		"Rule:Unique matched:%d, other:%d",
		ready.ID,
		other.ID)
	Log.Info(reason)
	return
}

// RuleDeps - Task kind dependencies.
type RuleDeps struct {
	cluster *Cluster
}

// Match determines the match.
func (r *RuleDeps) Match(ready, other *Task) (matched bool, reason string) {
	if ready.Kind == "" || other.Kind == "" {
		return
	}
	if *ready.ApplicationID != *other.ApplicationID {
		return
	}
	def, found := r.cluster.tasks[ready.Kind]
	if !found {
		return
	}
	matched = def.HasDep(other.Kind)
	reason = fmt.Sprintf(
		"Rule:Dependency matched:%d, other:%d",
		ready.ID,
		other.ID)
	Log.Info(reason)
	return
}

// RulePreempted - preempted tasks postponed to prevent thrashing.
type RulePreempted struct {
}

// Match determines the match.
// Postpone based on a duration after the last preempted event.
func (r *RulePreempted) Match(ready, _ *Task) (matched bool, reason string) {
	preemption := Settings.Hub.Task.Preemption
	if !preemption.Enabled {
		return
	}
	mark := time.Now()
	event, found := ready.LastEvent(Preempted)
	if found {
		if mark.Sub(event.Last) < preemption.Postponed {
			matched = true
			reason = fmt.Sprintf(
				"Rule:Preempted id:%d",
				ready.ID)
			Log.Info(reason)
		}
	}
	return
}

// RuleIsolated policy.
type RuleIsolated struct {
}

// Match determines the match.
func (r *RuleIsolated) Match(ready, other *Task) (matched bool, reason string) {
	matched = ready.Policy.Isolated || other.Policy.Isolated
	reason = fmt.Sprintf(
		"Rule:Isolated matched:%d, other:%d",
		ready.ID,
		other.ID)
	Log.Info(reason)
	return
}
