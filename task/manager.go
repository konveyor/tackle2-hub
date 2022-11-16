package task

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/auth"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

//
// States
const (
	Created   = "Created"
	Postponed = "Postponed"
	Ready     = "Ready"
	Pending   = "Pending"
	Running   = "Running"
	Succeeded = "Succeeded"
	Failed    = "Failed"
	Canceled  = "Canceled"
)

//
// Policies
const (
	Isolated = "isolated"
)

const (
	Unit = time.Second
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("task-scheduler")
)

//
// AddonNotFound used to report addon referenced
// by a task but cannot be found.
type AddonNotFound struct {
	Name string
}

func (e *AddonNotFound) Error() (s string) {
	return fmt.Sprintf("Addon: '%s' not-found.", e.Name)
}

func (e *AddonNotFound) Is(err error) (matched bool) {
	_, matched = err.(*AddonNotFound)
	return
}

//
// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
	// Addon token scopes.
	Scopes []string
}

//
// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	auth.Validators = append(
		auth.Validators,
		&Validator{
			Client: m.Client,
			DB:     m.DB,
		})
	go func() {
		Log.Info("Started.")
		defer Log.Info("Done.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.updateRunning()
				m.startReady()
				m.pause()
			}
		}
	}()
}

//
// Pause.
func (m *Manager) pause() {
	d := Unit * time.Duration(Settings.Frequency.Task)
	time.Sleep(d)
}

//
// startReady starts pending tasks.
func (m *Manager) startReady() {
	list := []model.Task{}
	db := m.DB.Order("priority DESC, id")
	result := db.Find(
		&list,
		"state IN ?",
		[]string{
			Ready,
			Postponed,
			Pending,
			Running,
		})
	Log.Trace(result.Error)
	if result.Error != nil {
		return
	}
	for i := range list {
		task := &list[i]
		if task.Canceled {
			m.canceled(task)
			continue
		}
		switch task.State {
		case Ready,
			Postponed:
			ready := task
			if m.postpone(ready, list) {
				ready.State = Postponed
				Log.Info("Task postponed.", "id", ready.ID)
				sErr := m.DB.Save(ready).Error
				Log.Trace(sErr)
				continue
			}
			rt := Task{ready}
			err := rt.Run(m.Client)
			if err != nil {
				if errors.Is(err, &AddonNotFound{}) {
					ready.Error = err.Error()
					ready.State = Failed
					sErr := m.DB.Save(ready).Error
					Log.Trace(sErr)
				}
				Log.Trace(err)
				continue
			}
			Log.Info("Task started.", "id", ready.ID)
			err = m.DB.Save(ready).Error
			Log.Trace(err)
		default:
			// Ignored.
			// Other states included to support
			// postpone rules.
		}
	}
}

//
// updateRunning tasks to reflect pod state.
func (m *Manager) updateRunning() {
	list := []model.Task{}
	db := m.DB.Order("priority DESC, id")
	result := db.Find(
		&list,
		"state IN ?",
		[]string{
			Pending,
			Running,
		})
	Log.Trace(result.Error)
	if result.Error != nil {
		return
	}
	for _, running := range list {
		if running.Canceled {
			m.canceled(&running)
			continue
		}
		rt := Task{&running}
		err := rt.Reflect(m.Client)
		if err != nil {
			Log.Trace(err)
			continue
		}
		err = m.DB.Save(&running).Error
		if err != nil {
			Log.Trace(result.Error)
			continue
		}
		Log.V(1).Info("Task updated.", "id", running.ID)
	}
}

//
// postpone Postpones a task as needed based on rules.
func (m *Manager) postpone(ready *model.Task, list []model.Task) (postponed bool) {
	ruleSet := []Rule{
		&RuleIsolated{},
		&RuleUnique{},
	}
	for i := range list {
		other := &list[i]
		if ready.ID == other.ID {
			continue
		}
		switch other.State {
		case Running,
			Pending:
			for _, rule := range ruleSet {
				if rule.Match(ready, other) {
					postponed = true
					return
				}
			}
		}
	}

	return
}

//
// The task has been canceled.
func (m *Manager) canceled(task *model.Task) {
	rt := Task{task}
	err := rt.Cancel(m.Client)
	Log.Trace(err)
	if err != nil {
		return
	}
	err = m.DB.Save(task).Error
	Log.Trace(err)
	db := m.DB.Model(&model.TaskReport{})
	err = db.Delete("taskid", task.ID).Error
	Log.Trace(err)
	return
}

//
// Task is an runtime task.
type Task struct {
	// model.
	*model.Task
}

//
// Run the specified task.
func (r *Task) Run(client k8s.Client) (err error) {
	mark := time.Now()
	defer func() {
		if err != nil {
			r.Error = err.Error()
			r.Terminated = &mark
			r.State = Failed
		}
	}()
	addon, err := r.findAddon(client, r.Addon)
	if err != nil {
		return
	}
	owner, err := r.findTackle(client)
	if err != nil {
		return
	}
	r.Image = addon.Spec.Image
	secret := r.secret(addon)
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
	pod := r.pod(addon, owner, &secret)
	err = client.Create(context.TODO(), &pod)
	if err != nil {
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
	r.Started = &mark
	r.State = Pending
	r.Pod = path.Join(
		pod.Namespace,
		pod.Name)
	return
}

//
// Reflect finds the associated pod and updates the task state.
func (r *Task) Reflect(client k8s.Client) (err error) {
	pod := &core.Pod{}
	err = client.Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: path.Dir(r.Pod),
			Name:      path.Base(r.Pod),
		},
		pod)
	if err != nil {
		if k8serr.IsNotFound(err) {
			err = r.Run(client)
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	mark := time.Now()
	status := pod.Status
	switch status.Phase {
	case core.PodRunning:
		r.State = Running
	case core.PodSucceeded:
		r.State = Succeeded
		r.Terminated = &mark
	case core.PodFailed:
		if r.Retries < Settings.Hub.Task.Retries {
			_ = client.Delete(context.TODO(), pod)
			r.Pod = ""
			r.Error = ""
			r.State = Ready
			r.Retries++
		} else {
			r.State = Failed
			r.Terminated = &mark
			r.Error = "pod failed."
		}
	}

	return
}

//
// Delete the associated pod as needed.
func (r *Task) Delete(client k8s.Client) (err error) {
	if r.Pod == "" {
		return
	}
	pod := &core.Pod{}
	pod.Namespace = path.Dir(r.Pod)
	pod.Name = path.Base(r.Pod)
	err = client.Delete(context.TODO(), pod)
	if err == nil {
		r.Pod = ""
		Log.Info(
			"Task pod deleted.",
			"id",
			r.ID,
			"pod",
			pod.Name)
	} else {
		err = liberr.Wrap(err)
		return
	}
	mark := time.Now()
	r.Terminated = &mark
	return
}

//
// Cancel the task.
func (r *Task) Cancel(client k8s.Client) (err error) {
	err = r.Delete(client)
	if err != nil {
		return
	}
	r.State = Canceled
	r.Bucket = ""
	Log.Info(
		"Task canceled.",
		"id",
		r.ID)
	return
}

//
// findAddon by name.
func (r *Task) findAddon(client k8s.Client, name string) (addon *crd.Addon, err error) {
	addon = &crd.Addon{}
	err = client.Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		addon)
	if err != nil {
		if k8serr.IsNotFound(err) {
			err = &AddonNotFound{name}
		} else {
			err = liberr.Wrap(err)
		}
		return
	}

	return
}

//
// findTackle returns the tackle CR.
func (r *Task) findTackle(client k8s.Client) (owner *crd.Tackle, err error) {
	list := crd.TackleList{}
	err = client.List(
		context.TODO(),
		&k8s.ListOptions{Namespace: Settings.Namespace},
		&list)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(list.Items) == 0 {
		err = liberr.New("Tackle CR not found.")
		return
	}
	owner = &list.Items[0]
	return
}

//
// pod build the pod.
func (r *Task) pod(addon *crd.Addon, owner *crd.Tackle, secret *core.Secret) (pod core.Pod) {
	pod = core.Pod{
		Spec: r.specification(addon, secret),
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

//
// specification builds a Pod specification.
func (r *Task) specification(addon *crd.Addon, secret *core.Secret) (specification core.PodSpec) {
	specification = core.PodSpec{
		ServiceAccountName: Settings.Hub.Task.SA,
		RestartPolicy:      core.RestartPolicyNever,
		Containers: []core.Container{
			r.container(addon, secret),
		},
		Volumes: []core.Volume{
			{
				Name: "working",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "bucket",
				VolumeSource: core.VolumeSource{
					PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
						ClaimName: Settings.Hub.Bucket.PVC,
					},
				},
			},
		},
	}
	mounts := addon.Spec.Mounts
	for _, mnt := range mounts {
		specification.Volumes = append(
			specification.Volumes,
			core.Volume{
				Name: mnt.Name,
				VolumeSource: core.VolumeSource{
					PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
						ClaimName: mnt.Claim,
					},
				},
			})
	}

	return
}

//
// container builds the pod container.
func (r *Task) container(addon *crd.Addon, secret *core.Secret) (container core.Container) {
	userid := int64(0)
	container = core.Container{
		Name:            "main",
		Image:           r.Image,
		ImagePullPolicy: core.PullAlways,
		WorkingDir:      Settings.Addon.Path.WorkingDir,
		Resources:       addon.Spec.Resources,
		Env: []core.EnvVar{
			{
				Name:  settings.EnvBucketPath,
				Value: Settings.Hub.Bucket.Path,
			},
			{
				Name:  settings.EnvHubBaseURL,
				Value: Settings.Addon.Hub.URL,
			},
			{
				Name:  settings.EnvAddonWorkingDir,
				Value: Settings.Addon.Path.WorkingDir,
			},
			{
				Name:  settings.EnvTask,
				Value: strconv.Itoa(int(r.Task.ID)),
			},
			{
				Name: settings.EnvHubToken,
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						Key: settings.EnvHubToken,
						LocalObjectReference: core.LocalObjectReference{
							Name: secret.Name,
						},
					},
				},
			},
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "working",
				MountPath: Settings.Addon.Path.WorkingDir,
			},
			{
				Name:      "bucket",
				MountPath: Settings.Hub.Bucket.Path,
			},
		},
		SecurityContext: &core.SecurityContext{
			RunAsUser: &userid,
		},
	}
	mounts := addon.Spec.Mounts
	for _, mnt := range mounts {
		container.VolumeMounts = append(
			container.VolumeMounts,
			core.VolumeMount{
				Name:      mnt.Name,
				MountPath: "/mnt/" + mnt.Name,
			},
		)
	}

	return
}

//
// secret builds the pod secret.
func (r *Task) secret(addon *crd.Addon) (secret core.Secret) {
	user := "addon:" + addon.Name
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

//
// k8sName returns a name suitable to be used for k8s resources.
func (r *Task) k8sName() string {
	return fmt.Sprintf("task-%d-", r.ID)
}

//
// labels builds k8s labels.
func (r *Task) labels() map[string]string {
	return map[string]string{
		"task": strconv.Itoa(int(r.ID)),
		"app":  "tackle",
		"role": "task",
	}
}
