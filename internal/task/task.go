package task

import (
	"context"
	"fmt"
	"hash/fnv"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/model/reflect"
	"github.com/konveyor/tackle2-hub/internal/ptr"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTask(m *model.Task) (t *Task) {
	t = &Task{}
	t.With(m)
	return
}

func NewTaskGroup(m *model.TaskGroup) (t *TaskGroup) {
	t = &TaskGroup{}
	t.With(m)
	return
}

// Task is an runtime task.
type Task struct {
	// model.
	*model.Task
	digest string
}

// With initializes the object.
func (r *Task) With(m *model.Task) {
	r.Task = m
	r.digest = r.getDigest()
}

// StateIn returns true matches on of the specified states.
func (r *Task) StateIn(states ...string) (matched bool) {
	for _, state := range states {
		if r.State == state {
			matched = true
			break
		}
	}
	return
}

// Error appends an error.
func (r *Task) Error(severity, description string, x ...any) {
	description = fmt.Sprintf(description, x...)
	r.Errors = append(
		r.Errors,
		model.TaskError{
			Severity:    severity,
			Description: description,
		})
}

// Event appends an event.
// Duplicates result in count incremented and Last updated.
func (r *Task) Event(kind string, p ...any) {
	mark := time.Now()
	reason := ""
	if len(p) > 0 {
		switch x := p[0].(type) {
		case string:
			reason = fmt.Sprintf(x, p[1:]...)
		case int:
			reason = strconv.Itoa(x)
		}
	}
	event, found := r.LastEvent(kind)
	if found && event.Reason == reason {
		event.Last = mark
		event.Count++
		return
	}
	event = &model.TaskEvent{
		Kind:   kind,
		Count:  1,
		Reason: reason,
		Last:   mark,
	}
	r.Events = append(r.Events, *event)
}

// LastEvent returns the last event of the specified kind.
func (r *Task) LastEvent(kind string) (event *model.TaskEvent, found bool) {
	for i := len(r.Events) - 1; i >= 0; i-- {
		event = &r.Events[i]
		if kind == event.Kind {
			found = true
			break
		}
	}
	return
}

// FindEvent returns the matched events by kind.
func (r *Task) FindEvent(kind string) (matched []*model.TaskEvent) {
	for i := 0; i < len(r.Events); i++ {
		event := &r.Events[i]
		if kind == event.Kind {
			matched = append(matched, event)
		}
	}
	return
}

// Run the specified task.
func (r *Task) Run(cluster *Cluster, quota *Quota) (started bool, err error) {
	mark := time.Now()
	client := cluster.Client
	defer func() {
		if err == nil {
			if started {
				quota.created()
			}
			return
		}
		matched, retry := SoftErr(err)
		if matched {
			if !retry {
				r.Error("Error", err.Error())
				r.Terminated = &mark
				r.State = Failed
			}
			err = nil
		}
	}()
	if quota.exhausted() {
		qe := &QuotaExceeded{Reason: quota.string()}
		r.State = QuotaBlocked
		r.Event(QuotaBlocked, qe.Reason)
		err = qe
		return
	}
	addon, found := cluster.Addon(r.Addon)
	if !found {
		err = &AddonNotFound{Name: r.Addon}
		return
	}
	extensions, err := cluster.FindExtensions(r.Extensions)
	if err != nil {
		return
	}
	for _, extension := range extensions {
		matched := false
		matched, err = r.matchAddon(extension, addon)
		if err != nil {
			return
		}
		if !matched {
			err = &ExtensionNotValid{
				Name:  extension.Name,
				Addon: addon.Name,
			}
			return
		}
	}
	secret := r.secret()
	err = client.Create(context.TODO(), &secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		if err != nil {
			_ = client.Delete(context.TODO(), &secret)
		}
	}()
	pod := r.pod(
		addon,
		extensions,
		cluster.Tackle(),
		&secret)
	err = client.Create(context.TODO(), &pod)
	if err != nil {
		qe := &QuotaExceeded{}
		if qe.Match(err) {
			r.State = QuotaBlocked
			r.Event(QuotaBlocked, qe.Reason)
			err = qe
			return
		}
		pe := &PodRejected{}
		if pe.Match(err) {
			err = liberr.Wrap(pe)
			return
		}
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		if err != nil {
			_ = client.Delete(context.TODO(), &pod)
		}
	}()
	secret.OwnerReferences = append(
		secret.OwnerReferences,
		meta.OwnerReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       pod.Name,
			UID:        pod.UID,
		})
	err = client.Update(context.TODO(), &secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	started = true
	r.Started = &mark
	r.State = Pending
	r.Pod = path.Join(
		pod.Namespace,
		pod.Name)
	r.Event(PodCreated, r.Pod)
	return
}

// Reflect finds the associated pod and updates the task state.
func (r *Task) Reflect(cluster *Cluster) (pod *core.Pod, found bool) {
	pod, found = cluster.Pod(path.Base(r.Pod))
	if !found {
		r.State = Ready
		r.Event(PodNotFound, r.Pod)
		r.Terminated = nil
		r.Started = nil
		r.Pod = ""
		return
	}
	client := cluster.Client
	switch pod.Status.Phase {
	case core.PodPending:
		r.podPending(pod)
	case core.PodRunning:
		r.podRunning(pod, client)
	case core.PodSucceeded:
		r.podSucceeded(pod)
	case core.PodFailed:
		r.podFailed(pod, client)
	}

	return
}

// Delete the associated pod as needed.
func (r *Task) Delete(client k8s.Client) (err error) {
	if r.Pod == "" {
		return
	}
	pod := &core.Pod{}
	pod.Namespace = path.Dir(r.Pod)
	pod.Name = path.Base(r.Pod)
	err = client.Delete(context.TODO(), pod, k8s.GracePeriodSeconds(0))
	if err != nil {
		if !k8serr.IsNotFound(err) {
			err = liberr.Wrap(err)
			return
		} else {
			err = nil
		}
	}
	r.Pod = ""
	r.Event(PodDeleted, r.Pod)
	Log.Info(
		"Task pod deleted.",
		"id",
		r.ID,
		"pod",
		pod.Name)
	mark := time.Now()
	r.Terminated = &mark
	return
}

// MatchSubject returns true when the other task has the same subject.
func (r *Task) MatchSubject(other *Task) (matched bool) {
	id := ptr.ID(r.ApplicationID)
	otherId := ptr.ID(other.ApplicationID)
	matched = id != 0 && otherId != 0 && id == otherId
	if !matched {
		id = ptr.ID(r.PlatformID)
		otherId = ptr.ID(other.PlatformID)
		matched = id != 0 && otherId != 0 && id == otherId
	}
	return
}

// podPending handles pod pending.
func (r *Task) podPending(pod *core.Pod) {
	var status []core.ContainerStatus
	status = append(
		status,
		pod.Status.InitContainerStatuses...)
	status = append(
		status,
		pod.Status.ContainerStatuses...)
	started := 0
	for _, cnd := range pod.Status.Conditions {
		if cnd.Type == core.PodScheduled &&
			cnd.Reason == core.PodReasonUnschedulable {
			r.Event(PodUnschedulable, cnd.Message)
			return
		}
	}
	for _, status := range status {
		state := status.State
		if state.Waiting != nil {
			waiting := state.Waiting
			reason := strings.ToLower(waiting.Reason)
			if r.containsAny(reason, "invalid", "error", "never", "cannot") {
				r.Error(
					"Error",
					"Container (%s) failed: %s",
					status.Name,
					waiting.Reason)
				mark := time.Now()
				r.Terminated = &mark
				r.Event(ImageError, waiting.Reason)
				r.State = Failed
				return
			} else {
				r.Event(PodPending, waiting.Reason)
			}
		}
		if status.Started == nil {
			continue
		}
		if *status.Started {
			started++
		}
	}
	if started > 0 {
		r.Event(PodRunning)
		r.State = Running
	}
}

// Cancel the task.
func (r *Task) Cancel(client k8s.Client) (err error) {
	err = r.Delete(client)
	if err != nil {
		return
	}
	r.State = Canceled
	r.Event(Canceled)
	r.SetBucket(nil)
	Log.Info(
		"Task canceled.",
		"id",
		r.ID)
	return
}

// podRunning handles pod running.
// container[0] is the addon.
// succeeded when: the addon has terminated with exit=0.
// failed when: any container has terminated != 0.
func (r *Task) podRunning(pod *core.Pod, client k8s.Client) {
	r.State = Running
	r.Event(PodRunning)
	for i := range pod.Status.ContainerStatuses {
		status := &pod.Status.ContainerStatuses[i]
		if status.State.Terminated == nil {
			continue
		}
		switch status.State.Terminated.ExitCode {
		case 0:
			if i == 0 { // addon
				r.podSucceeded(pod)
			}
		default: // failed.
			r.podFailed(pod, client)
			return
		}
	}
}

// podFailed handles pod succeeded.
func (r *Task) podSucceeded(pod *core.Pod) {
	mark := time.Now()
	r.State = Succeeded
	r.Event(PodSucceeded)
	r.Terminated = &mark
}

// podFailed handles pod failed.
func (r *Task) podFailed(pod *core.Pod, client k8s.Client) {
	mark := time.Now()
	var statuses []core.ContainerStatus
	statuses = append(
		statuses,
		pod.Status.InitContainerStatuses...)
	statuses = append(
		statuses,
		pod.Status.ContainerStatuses...)
	if len(statuses) == 0 {
		r.State = Failed
		r.Terminated = &mark
		r.Error(
			"Error",
			"Pod (%s) failed: %s",
			pod.Name,
			pod.Status.Reason)
		return
	}
	for _, status := range statuses {
		if status.State.Terminated == nil {
			continue
		}
		r.Event(
			PodFailed,
			status.State.Terminated.Reason)
		switch status.State.Terminated.ExitCode {
		case 0: // Succeeded.
		case 137: // Killed.
			if r.Retries < Settings.Hub.Task.Retries {
				_ = client.Delete(context.TODO(), pod)
				r.Pod = ""
				r.State = Ready
				r.Terminated = nil
				r.Started = nil
				r.Errors = nil
				r.Retries++
				return
			}
			fallthrough
		default: // Error.
			r.State = Failed
			r.Terminated = &mark
			r.Error(
				"Error",
				"Container (%s) failed: %s",
				status.Name,
				status.State.Terminated.Reason)
			return
		}
	}
}

// matchTask - returns true when the addon's `task`
// (ref) matches the task name.
// The `ref` is matched as a REGEX when it contains
// characters other than: [0-9A-Za-z_].
func (r *Task) matchTask(addon *crd.Addon, task *crd.Task) (matched bool, err error) {
	ref := strings.TrimSpace(addon.Spec.Task)
	p := IsRegex
	if p.MatchString(ref) {
		p, err = regexp.Compile(ref)
		if err != nil {
			err = &AddonTaskNotValid{
				Addon:  addon.Name,
				Reason: err.Error(),
			}
			return
		}
		matched = p.MatchString(task.Name)
	} else {
		matched = task.Name == ref
	}
	return
}

// matchAddon - returns true when the extension's `addon`
// (ref) matches the addon name.
// The `ref` is matched as a REGEX when it contains
// characters other than: [0-9A-Za-z_].
func (r *Task) matchAddon(extension *crd.Extension, addon *crd.Addon) (matched bool, err error) {
	ref := strings.TrimSpace(extension.Spec.Addon)
	p := IsRegex
	if p.MatchString(ref) {
		p, err = regexp.Compile(ref)
		if err != nil {
			err = &ExtAddonNotValid{
				Extension: extension.Name,
				Reason:    err.Error(),
			}
			return
		}
		matched = p.MatchString(addon.Name)
	} else {
		matched = addon.Name == ref
	}
	return
}

// pod build the pod.
func (r *Task) pod(
	addon *crd.Addon,
	extensions []*crd.Extension,
	owner *crd.Tackle,
	secret *core.Secret) (pod core.Pod) {
	//
	pod = core.Pod{
		Spec: r.specification(addon, extensions, secret),
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: r.k8sName(),
			Labels:       r.labels(),
		},
	}
	pod.OwnerReferences = append(
		pod.OwnerReferences,
		meta.OwnerReference{
			APIVersion: owner.APIVersion,
			Kind:       owner.Kind,
			Name:       owner.Name,
			UID:        owner.UID,
		})
	return
}

// specification builds a Pod specification.
func (r *Task) specification(
	addon *crd.Addon,
	extensions []*crd.Extension,
	secret *core.Secret) (specification core.PodSpec) {
	addonDir := core.Volume{
		Name: Addon,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	shared := core.Volume{
		Name: Shared,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
	cache := core.Volume{
		Name: Cache,
	}
	if Settings.Cache.RWX {
		cache.VolumeSource = core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: Settings.Cache.PVC,
			},
		}
	} else {
		cache.VolumeSource = core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		}
	}
	init, plain := r.containers(addon, extensions, secret)
	specification = core.PodSpec{
		ServiceAccountName: Settings.Hub.Task.SA,
		RestartPolicy:      core.RestartPolicyNever,
		InitContainers:     init,
		Containers:         plain,
		Volumes: []core.Volume{
			addonDir,
			shared,
			cache,
		},
	}

	return
}

// container builds the pod containers.
func (r *Task) containers(
	addon *crd.Addon,
	extensions []*crd.Extension,
	secret *core.Secret) (init []core.Container, plain []core.Container) {
	token := &core.EnvVarSource{
		SecretKeyRef: &core.SecretKeySelector{
			Key: settings.EnvHubToken,
			LocalObjectReference: core.LocalObjectReference{
				Name: secret.Name,
			},
		},
	}
	uid := Settings.Hub.Task.UID
	plain = append(plain, addon.Spec.Container)
	plain[0].Name = "addon"
	for i := range extensions {
		extension := extensions[i]
		container := extension.Spec.Container
		container.Name = extension.Name
		plain = append(
			plain,
			container)
	}
	injector := Injector{}
	for i := range plain {
		container := &plain[i]
		injector.Inject(container)
		r.propagateEnv(&plain[0], container)
		container.SecurityContext = &core.SecurityContext{
			RunAsUser: &uid,
		}
		container.VolumeMounts = append(
			container.VolumeMounts,
			core.VolumeMount{
				Name:      Addon,
				MountPath: Settings.Addon.HomeDir,
			},
			core.VolumeMount{
				Name:      Shared,
				MountPath: Settings.Addon.SharedDir,
			},
			core.VolumeMount{
				Name:      Cache,
				MountPath: Settings.Addon.CacheDir,
			})
		container.Env = append(
			container.Env,
			core.EnvVar{
				Name:  settings.EnvAddonHomeDir,
				Value: Settings.Addon.HomeDir,
			},
			core.EnvVar{
				Name:  settings.EnvSharedDir,
				Value: Settings.Addon.SharedDir,
			},
			core.EnvVar{
				Name:  settings.EnvCacheDir,
				Value: Settings.Addon.CacheDir,
			},
			core.EnvVar{
				Name:  settings.EnvHubBaseURL,
				Value: Settings.Addon.Hub.URL,
			},
			core.EnvVar{
				Name:  settings.EnvTask,
				Value: strconv.Itoa(int(r.Task.ID)),
			},
			core.EnvVar{
				Name:      settings.EnvHubToken,
				ValueFrom: token,
			})
	}
	return
}

// propagateEnv copies extension container Env.* to the addon container.
// Prefixed with EXTENSION_<name>.
func (r *Task) propagateEnv(addon, extension *core.Container) {
	for _, env := range extension.Env {
		addon.Env = append(
			addon.Env,
			core.EnvVar{
				Name:  ExtEnv(extension.Name, env.Name),
				Value: env.Value,
			})
	}
}

// secret builds the pod secret.
func (r *Task) secret() (secret core.Secret) {
	user := "addon:" + r.Addon
	token, _ := auth.Hub.NewToken(
		user,
		auth.AddonRole,
		jwt.MapClaims{
			"task": r.ID,
		})
	secret = core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: r.k8sName(),
			Labels:       r.labels(),
		},
		Data: map[string][]byte{
			settings.EnvHubToken: []byte(token),
		},
	}

	return
}

// k8sName returns a name suitable to be used for k8s resources.
func (r *Task) k8sName() string {
	return fmt.Sprintf("task-%d-", r.ID)
}

// labels builds k8s labels.
func (r *Task) labels() map[string]string {
	return map[string]string{
		TaskLabel: strconv.Itoa(int(r.ID)),
		AppLabel:  "tackle",
		RoleLabel: "task",
	}
}

// attach file.
func (r *Task) attach(file *model.File) {
	r.Attached = append(
		r.Attached,
		model.Attachment{
			ID:   file.ID,
			Name: file.Name,
		})
}

// containsAny returns true when the str contains any of substr.
func (r *Task) containsAny(str string, substr ...string) (matched bool) {
	for i := range substr {
		if strings.Contains(str, substr[i]) {
			matched = true
			break
		}
	}
	return
}

// update manager controlled fields.
func (r *Task) update(db *gorm.DB) (err error) {
	db = reflect.Select(
		db,
		r.Task,
		"Addon",
		"Extensions",
		"State",
		"Priority",
		"Started",
		"Terminated",
		"Retained",
		"Events",
		"Errors",
		"Retries",
		"Attached",
		"Pod")
	err = db.Save(r).Error
	if err == nil {
		Log.V(1).Info("Task updated.", "id", r.ID)
		r.With(r.Task)
	}
	return
}

// getDigest returns a digest.
func (r *Task) getDigest() (d string) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(r.Addon))
	_, _ = h.Write([]byte(fmt.Sprintf("%#v", r.Extensions)))
	_, _ = h.Write([]byte(r.State))
	_, _ = h.Write([]byte(strconv.Itoa(r.Priority)))
	_, _ = h.Write([]byte(fmt.Sprintf("%#v", r.Events)))
	_, _ = h.Write([]byte(fmt.Sprintf("%#v", r.Errors)))
	_, _ = h.Write([]byte(strconv.Itoa(r.Retries)))
	_, _ = h.Write([]byte(fmt.Sprintf("%#v", r.Attached)))
	_, _ = h.Write([]byte(r.Pod))
	if r.Started != nil {
		_, _ = h.Write([]byte(r.Started.String()))
	}
	if r.Terminated != nil {
		_, _ = h.Write([]byte(r.Terminated.String()))
	}
	if r.Retained {
		_, _ = h.Write([]byte{1})
	}
	n := h.Sum64()
	d = fmt.Sprintf("%x", n)
	return
}

// hasChanged returns true when the digest has changed.
func (r *Task) hasChanged() (changed bool) {
	d := r.getDigest()
	changed = d != r.digest
	return
}

// podRetentionExpired returns true when the retention period has expired.
func (r *Task) podRetentionExpired() (expired bool) {
	period := r.podRetention()
	mark := r.CreateTime
	if r.Terminated != nil {
		mark = *r.Terminated
	}
	d := time.Duration(period) * time.Second
	expired = time.Since(mark) > d
	return
}

// retention returns the retention period (seconds).
func (r *Task) podRetention() (seconds int) {
	if r.State == Succeeded {
		seconds = Settings.Hub.Task.Pod.Retention.Succeeded
	} else {
		seconds = Settings.Hub.Task.Pod.Retention.Failed
	}
	return
}

// Event represents a pod event.
type Event struct {
	Type     string
	Reason   string
	Age      string
	Reporter string
	Message  string
}

// TaskGroup represents a task group.
type TaskGroup struct {
	*model.TaskGroup
}

// Submit the task group.
// - propagate properties to members.
// - create member (tasks).
func (g *TaskGroup) Submit(db *gorm.DB, manager *Manager) (err error) {
	g.State = Ready
	err = g.propagate()
	if err != nil {
		return
	}
	gdb := db.Omit(clause.Associations)
	err = gdb.Save(g).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range g.Tasks {
		task := &Task{}
		task.With(&g.Tasks[i])
		if task.ID > 0 {
			err = &BadRequest{
				Reason: "tasks already created",
			}
			return
		}
		task.TaskGroupID = &g.ID
		err = manager.Create(db, task)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

func (g *TaskGroup) With(m *model.TaskGroup) {
	g.TaskGroup = m
}

// propagate group data into the task.
func (g *TaskGroup) propagate() (err error) {
	m := g.TaskGroup
	m.Tasks = make([]model.Task, 0)
	for i := range m.List {
		m.Tasks = append(
			m.Tasks,
			m.List[i])
	}
	for i := range m.Tasks {
		task := &m.Tasks[i]
		task.CreateUser = m.CreateUser
		task.UpdateUser = m.UpdateUser
		switch m.Mode {
		case "", Batch:
			task.State = m.State
			task.Kind = m.Kind
			task.Addon = m.Addon
			task.Extensions = m.Extensions
			task.Priority = m.Priority
			task.Policy = m.Policy
			task.SetBucket(m.BucketID)
			merged := task.Data.Merge(m.Data)
			if !merged {
				task.Data = m.Data
			}
		case Pipeline:
			if i == 0 {
				task.State = m.State
				task.SetBucket(m.BucketID)
			}
		}
	}

	return
}
