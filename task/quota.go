package task

import (
	"fmt"
)

// Quota tracks task pod quota/capacity.
type Quota struct {
	quota    int
	count    int
	capacity int
}

// with init with cluster.
// quota Zero(0) is unlimited.
func (q *Quota) with(k *Cluster) {
	q.quota = Settings.Hub.Task.Pod.Quota
	q.count = 0
	q.capacity = 0
	if qty, found := k.PodQuota(); found {
		q.quota = qty
		q.quota -= len(k.OtherPods())
		q.quota = max(q.quota, 0)
	}
	q.count = len(k.TaskPods())
	q.capacity = q.quota - q.count
}

// created indicates a task pod has been created.
// increments count; decrements the capacity.
func (q *Quota) created() {
	q.capacity--
	q.count++
}

// exhausted returns true when the capacity < 1.
// A zero(0) quota is unlimited.
func (q *Quota) exhausted() (exhausted bool) {
	if q.quota < 1 {
		return
	}
	exhausted = q.capacity < 1
	return
}

// string returns a string representation.
func (q *Quota) string() (s string) {
	if q.quota > 0 {
		s = fmt.Sprintf(
			"quota (pod): %d/%d",
			q.count,
			q.quota)
	} else {
		s = fmt.Sprintf(
			"quota (pod): %d/*",
			q.count)
	}
	return s
}
