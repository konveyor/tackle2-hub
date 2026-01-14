package task

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/internal/model"
)

// Rule defines postpone rules.
type Rule interface {
	Match(*Task, *Domain) (matched bool, reason string)
}

// RuleUnique running tasks must be unique by:
//   - application
//   - addon.
type RuleUnique struct {
	matched map[uint]uint
}

// Match determines the match.
// Match on:
// - task (when specified on both)
// - addon
// - subject
func (r *RuleUnique) Match(ready *Task, d *Domain) (matched bool, reason string) {
	var other *Task
	if ready.Kind != "" {
		matched, other = d.matchKind(ready)
	}
	if !matched {
		matched, other = d.matchAddon(ready)
	}
	if !matched {
		return
	}
	if _, found := r.matched[other.ID]; found {
		matched = false
		return
	}
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
func (r *RuleDeps) Match(ready *Task, d *Domain) (matched bool, reason string) {
	if ready.Kind == "" {
		return
	}
	def, found := r.cluster.Task(ready.Kind)
	if !found {
		return
	}
	for _, kind := range def.Deps() {
		other := NewTask(&model.Task{})
		other.Kind = kind
		other.ApplicationID = ready.ApplicationID
		other.PlatformID = ready.PlatformID
		matched, other = d.matchKind(other)
		if matched {
			reason = fmt.Sprintf(
				"Rule:Dependency matched:%d, other:%d",
				ready.ID,
				other.ID)
			Log.Info(reason)
			break
		}
	}
	return
}

// RuleIsolated policy.
type RuleIsolated struct {
}

// Match determines the match.
func (r *RuleIsolated) Match(ready *Task, d *Domain) (matched bool, reason string) {
	matched, other := d.matchIsolated(ready)
	if matched {
		reason = fmt.Sprintf(
			"Rule:Isolated matched:%d, other:%d",
			ready.ID,
			other.ID)
		Log.Info(reason)
	}
	return
}

// Domain of tasks being scheduled.
type Domain struct {
	tasks      []*Task
	byKind     map[string][]*Task
	byAddon    map[string][]*Task
	byIsolated []*Task
}

// Load with tasks.
func (d *Domain) Load(list []*Task) {
	d.tasks = list
	d.byKind = make(map[string][]*Task)
	d.byAddon = make(map[string][]*Task)
	for _, task := range list {
		if task.Kind != "" {
			key := d.kind(task)
			d.byKind[key] = append(
				d.byKind[key],
				task)
		}
		if task.Addon != "" {
			key := d.addon(task)
			d.byAddon[key] = append(
				d.byAddon[key],
				task)
		}
		if task.Policy.Isolated {
			d.byIsolated = append(
				d.byIsolated,
				task)
		}
	}
}

// matchKind matches tasks in the domain when:
// - different task id
// - same subject
// - kind specified
// - kind matched
func (d *Domain) matchKind(task *Task) (found bool, matched *Task) {
	if task.Kind == "" {
		return
	}
	for _, t := range d.byKind[d.kind(task)] {
		if t.ID != task.ID {
			matched = t
			found = true
			break
		}
	}
	return
}

// matchAddon matches tasks in the domain when:
// - different task id
// - same subject
// - addon specified
// - addon matched
func (d *Domain) matchAddon(task *Task) (found bool, matched *Task) {
	if task.Addon == "" {
		return
	}
	for _, t := range d.byAddon[d.addon(task)] {
		if t.ID != task.ID {
			matched = t
			found = true
			break
		}
	}
	return
}

// matchIsolated matches tasks in the domain when:
// - different task id
// - Policy.Isolated = True.
func (d *Domain) matchIsolated(task *Task) (found bool, matched *Task) {
	for _, t := range d.byIsolated {
		if t.ID != task.ID {
			matched = t
			found = true
			break
		}
	}
	return
}

// subject returns the subject key.
// format: (A|P):id.
func (d *Domain) subject(task *Task) (key string) {
	if task.ApplicationID != nil {
		key = fmt.Sprintf("A:%d", *task.ApplicationID)
		return
	}
	if task.PlatformID != nil {
		key = fmt.Sprintf("P:%d", *task.PlatformID)
		return
	}
	return
}

// kind returns the kind key.
// format: subject:kind
func (d *Domain) kind(task *Task) (kind string) {
	kind = d.subject(task) + ":" + task.Kind
	return
}

// addon returns the addon key.
// format: subject:addon
func (d *Domain) addon(task *Task) (kind string) {
	kind = d.subject(task) + ":" + task.Addon
	return
}

func NewDomain(tasks []*Task) (d *Domain) {
	d = &Domain{}
	d.Load(tasks)
	return
}
