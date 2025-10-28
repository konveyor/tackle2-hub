package task

import (
	"fmt"
)

// Quota tracks task pod quota/capacity.
type Quota struct {
	quota int
	count int
}

// with updates with the cluster.
func (q *Quota) with(k *Cluster) {
	q.count = len(k.TaskPods())
	q.quota = Settings.Hub.Task.Pod.Quota
	if qty, found := k.PodQuota(); found {
		q.quota = qty
		nOther := len(k.OtherPods())
		q.quota -= nOther
	}
}

// created indicates a task pod has been created.
func (q *Quota) created() {
	q.count++
}

// exhausted returns true when the quota is reached.
func (q *Quota) exhausted() (exhausted bool) {
	exhausted = q.count >= q.quota
	return
}

// string returns a string representation.
func (q *Quota) string() (s string) {
	s = fmt.Sprintf(
		"quota (pod): %d/%d",
		q.count,
		q.quota)
	return s
}
