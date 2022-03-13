package task

import (
	"context"
	"encoding/json"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

const (
	Created   = "Created"
	Ready     = "Ready"
	Succeeded = "Succeeded"
	Failed    = "Failed"
	Running   = "Running"
	Postponed = "Postponed"
)

var Settings = &settings.Settings

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
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				_ = m.updateRunning()
				_ = m.purgeTerminated()
				_ = m.startReady()
			}
		}
	}()
}

//
// startReady starts pending tasks.
func (m *Manager) startReady() (err error) {
	list := []model.Task{}
	result := m.DB.Find(
		&list,
		"status IN ?",
		[]string{
			Ready,
			Running,
			Postponed,
		})
	if result.Error != nil {
		err = result.Error
		return
	}
	for i := range list {
		ready := &list[i]
		task := Task{
			client: m.Client,
			Task:   ready,
		}
		switch ready.Status {
		case Ready,
			Postponed:
			if m.postpone(ready, list) {
				ready.Status = Postponed
				_ = m.DB.Save(ready)
				continue
			}
			_ = task.Run()
			_ = m.DB.Save(ready)
		}
	}

	return
}

//
// updateRunning tasks to reflect job status.
func (m *Manager) updateRunning() (err error) {
	list := []model.Task{}
	result := m.DB.Find(&list, "status", Running)
	if result.Error != nil {
		err = result.Error
		return
	}
	for _, running := range list {
		task := Task{
			client: m.Client,
			Task:   &running,
		}
		err := task.Reflect()
		if err != nil {
			continue
		}
		_ = m.DB.Save(&running)
	}

	return
}

//
// purgeTerminated purge resources associated with terminated tasks.
//   - delete buckets.
func (m *Manager) purgeTerminated() (err error) {
	list := []model.Task{}
	result := m.DB.Find(
		&list,
		"status IN ?",
		[]string{
			Succeeded,
			Failed,
		})
	if result.Error != nil {
		err = result.Error
		return
	}
	for _, task := range list {
		delete := false
		done := time.Since(*task.Terminated)
		switch task.Status {
		case Succeeded:
			delete = done > time.Hour
		case Failed:
			delete = done > time.Hour*48
		}
		if delete {
			_ = os.Remove(task.Path)
		}
	}

	return
}

//
// postpone task based on requested isolation.
// An isolated task must run by itself and will cause all
// other tasks to be postponed.
func (m *Manager) postpone(pending *model.Task, list []model.Task) (found bool) {
	for i := range list {
		task := &list[i]
		if pending.ID == task.ID {
			continue
		}
		if pending.Status != Running {
			continue
		}
		if pending.Isolated || task.Isolated {
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
			r.Status = Failed
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
		return
	}
	job := r.job(&secret)
	err = r.client.Create(context.TODO(), &job)
	if err != nil {
		return
	}
	r.Started = &mark
	r.Status = Running
	r.Job = path.Join(
		job.Namespace,
		job.Name)
	return
}

//
// Reflect finds the associated job and updates the task status.
func (r *Task) Reflect() (err error) {
	job := &batch.Job{}
	err = r.client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: path.Dir(r.Job),
			Name:      path.Base(r.Job),
		},
		job)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Run()
		}
		return
	}
	mark := time.Now()
	status := job.Status
	for _, cnd := range status.Conditions {
		if cnd.Type == batch.JobFailed {
			r.Status = Failed
			r.Terminated = &mark
			r.Error = "job failed."
			return
		}
		if status.Succeeded > 0 {
			r.Status = Succeeded
			r.Terminated = &mark
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
		return
	}

	return
}

//
// job build the Job.
func (r *Task) job(secret *core.Secret) (job batch.Job) {
	template := r.template(secret)
	backOff := int32(2)
	job = batch.Job{
		Spec: batch.JobSpec{
			Template:     template,
			BackoffLimit: &backOff,
		},
		ObjectMeta: meta.ObjectMeta{
			Namespace:    Settings.Hub.Namespace,
			GenerateName: strings.ToLower(r.Name) + "-",
			Labels:       r.labels(),
		},
	}

	return
}

//
// template builds a Job template.
func (r *Task) template(secret *core.Secret) (template core.PodTemplateSpec) {
	template = core.PodTemplateSpec{
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
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
		},
	}
	mounts := r.addon.Spec.Mounts
	for _, mnt := range mounts {
		template.Spec.Volumes = append(
			template.Spec.Volumes,
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
// container builds the job container.
func (r *Task) container() (container core.Container) {
	container = core.Container{
		Name:       "main",
		Image:      r.Image,
		WorkingDir: Settings.Addon.Path.WorkingDir,
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
// secret builds the job secret.
func (r *Task) secret() (secret core.Secret) {
	data := Secret{}
	data.Hub.Task = r.Task.ID
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
		Token      string
		Task       uint
		Encryption struct {
			Passphrase string
		}
	}
	Addon interface{}
}
