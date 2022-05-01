package tasking

import (
	"context"
	"encoding/json"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
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
)

//
// Policies
const (
	Isolated = "Isolated"
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("task")
)

//
// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client client.Client
}

//
// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	scheduler := func() {
		Log.Info("Scheduler started.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				m.updateRunning()
				m.startReady()
			}
		}
	}
	reaper := func() {
		reapers := []Reaper{
			&TaskReaper{
				Client: m.Client,
				DB:     m.DB,
			},
			&GroupReaper{
				DB: m.DB,
			},
			&BucketReaper{
				DB: m.DB,
			},
		}
		Log.Info("Reaper started.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(ReaperUnit)
				for _, r := range reapers {
					r.Run()
				}
			}
		}
	}
	go scheduler()
	go reaper()
}

//
// startReady starts pending tasks.
func (m *Manager) startReady() {
	list := []model.Task{}
	result := m.DB.Find(
		&list,
		"state IN ?",
		[]string{
			Ready,
			Running,
			Postponed,
		})
	Log.Trace(result.Error)
	if result.Error != nil {
		return
	}
	for i := range list {
		ready := &list[i]
		task := Task{
			client: m.Client,
			Task:   ready,
		}
		switch ready.State {
		case Ready,
			Postponed:
			if m.postpone(ready, list) {
				ready.State = Postponed
				Log.Info("Task postponed.", "id", ready.ID)
			} else {
				err := task.Run()
				if err == nil {
					Log.Info("Task started.", "id", ready.ID)
				} else {
					continue
				}
			}
			result := m.DB.Save(ready)
			Log.Trace(result.Error)
		}
	}
}

//
// updateRunning tasks to reflect pod state.
func (m *Manager) updateRunning() {
	list := []model.Task{}
	result := m.DB.Find(
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
		task := Task{
			client: m.Client,
			Task:   &running,
		}
		err := task.Reflect()
		Log.Trace(err)
		if err != nil {
			continue
		}
		result := m.DB.Save(&running)
		Log.Trace(result.Error)
		if result.Error != nil {
			continue
		}
		Log.V(1).Info("Task updated.", "id", running.ID)
	}
}

//
// postpone Postpones a task based on the following rules:
//   - Tasks must be unique for an addon and application.
//   - An isolated task must run by itself and will cause all
//     other tasks to be postponed.
func (m *Manager) postpone(pending *model.Task, list []model.Task) (found bool) {
	for i := range list {
		task := &list[i]
		if task.ID == pending.ID {
			continue
		}
		if task.State != Running {
			continue
		}
		if pending.ApplicationID == task.ApplicationID && pending.Addon == task.Addon {
			found = true
			return
		}
		if pending.Policy == Isolated || task.Policy == Isolated {
			found = true
			return
		}
	}

	return
}

//
// Task is an runtime task.
type Task struct {
	// model.
	*model.Task
	// k8s client.
	client client.Client
	// addon
	addon *crd.Addon
}

//
// Run the specified task.
func (r *Task) Run() (err error) {
	mark := time.Now()
	defer func() {
		if err != nil {
			r.Error = err.Error()
			r.Terminated = &mark
			r.State = Failed
		}
	}()
	r.addon, err = r.findAddon(r.Addon)
	if err != nil {
		return
	}
	r.Image = r.addon.Spec.Image
	secret := r.secret()
	err = r.client.Create(context.TODO(), &secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	pod := r.pod(&secret)
	err = r.client.Create(context.TODO(), &pod)
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
func (r *Task) Reflect() (err error) {
	pod := &core.Pod{}
	err = r.client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: path.Dir(r.Pod),
			Name:      path.Base(r.Pod),
		},
		pod)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Run()
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
			_ = r.client.Delete(context.TODO(), pod)
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
// findAddon by name.
func (r *Task) findAddon(name string) (addon *crd.Addon, err error) {
	addon = &crd.Addon{}
	err = r.client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		addon)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

//
// pod build the pod.
func (r *Task) pod(secret *core.Secret) (pod core.Pod) {
	pod = core.Pod{
		Spec: r.specification(secret),
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: fmt.Sprintf("task-%d-", r.ID),
			Labels:       r.labels(),
		},
	}

	return
}

//
// specification builds a Pod specification.
func (r *Task) specification(secret *core.Secret) (specification core.PodSpec) {
	specification = core.PodSpec{
		ServiceAccountName: Settings.Hub.Task.SA,
		RestartPolicy:      core.RestartPolicyNever,
		Containers: []core.Container{
			r.container(),
		},
		Volumes: []core.Volume{
			{
				Name: "working",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "secret",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secret.Name,
					},
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
	mounts := r.addon.Spec.Mounts
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
func (r *Task) container() (container core.Container) {
	container = core.Container{
		Name:            "main",
		Image:           r.Image,
		ImagePullPolicy: core.PullAlways,
		WorkingDir:      Settings.Addon.Path.WorkingDir,
		Resources:       r.addon.Spec.Resources,
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
				Name:  settings.EnvAddonSecretPath,
				Value: Settings.Addon.Path.Secret,
			},
			{
				Name:  settings.EnvAddonWorkingDir,
				Value: Settings.Addon.Path.WorkingDir,
			},
		},
		VolumeMounts: []core.VolumeMount{
			{
				Name:      "working",
				MountPath: Settings.Addon.Path.WorkingDir,
			},
			{
				Name:      "secret",
				MountPath: path.Dir(Settings.Addon.Path.Secret),
			},
			{
				Name:      "bucket",
				MountPath: Settings.Hub.Bucket.Path,
			},
		},
	}
	mounts := r.addon.Spec.Mounts
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
func (r *Task) secret() (secret core.Secret) {
	data := Secret{}
	data.Hub.Task = r.Task.ID
	data.Hub.Application = r.Task.ApplicationID
	data.Hub.Encryption.Passphrase = Settings.Encryption.Passphrase
	data.Hub.Token = Settings.Auth.AddonToken
	data.Addon = r.Task.Data
	encoded, _ := json.Marshal(data)
	secret = core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: strings.ToLower(r.Name) + "-",
			Labels:       r.labels(),
		},
		Data: map[string][]byte{
			path.Base(Settings.Addon.Path.Secret): encoded,
		},
	}

	return
}

//
// labels builds k8s labels.
func (r *Task) labels() map[string]string {
	return map[string]string{
		"Task": strconv.Itoa(int(r.ID)),
	}
}

//
// Secret payload.
type Secret struct {
	Hub struct {
		Token       string
		Application *uint
		Task        uint
		Variant     string
		Encryption  struct {
			Passphrase string
		}
	}
	Addon interface{}
}
