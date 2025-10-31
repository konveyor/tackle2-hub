package task

import (
	"fmt"
)

// Quota tracks task pod quota/capacity.
type Quota struct {
	enabled bool
	quota   int
	count   int
}

// with updates with the cluster.
func (q *Quota) with(k *Cluster) {
	q.count = len(k.TaskPods())
	if qty, found := k.PodQuota(); found {
		nOther := len(k.OtherPods())
		q.quota = qty
		q.quota -= nOther
	} else {
		q.quota = Settings.Hub.Task.Pod.Quota
	}
	q.enabled = q.quota > 0
	q.quota = max(q.quota, 0)
}

// created indicates a task pod has been created.
func (q *Quota) created() {
	q.count++
}

// exhausted returns true when the quota is reached.
func (q *Quota) exhausted() (exhausted bool) {
	if !q.enabled {
		return
	}
	exhausted = q.count >= q.quota
	return
}

// string returns a string representation.
func (q *Quota) string() (s string) {
	if q.enabled {
		s = fmt.Sprintf(
			"%d/%d",
			q.count,
			q.quota)
	} else {
		s = "-/-"
	}
	return s
}
