package tasking

import (
	"context"
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				DB: m.DB,
			},
			&TaskReaper{
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
				result := m.DB.Save(ready)
				Log.Trace(result.Error)
				if result.Error == nil {
					Log.Info("Task postponed.", "id", ready.ID)
				}
			}
			err := task.Run()
			if err != nil {
				continue
			}
			Log.Info("Task started.", "id", ready.ID)
			result := m.DB.Save(ready)
			Log.Trace(result.Error)
		}
	}
}

//
// updateRunning tasks to reflect job state.
func (m *Manager) updateRunning() {
	list := []model.Task{}
	result := m.DB.Find(&list, "state", Running)
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
// postpone task based on requested isolation.
// An isolated task must run by itself and will cause all
// other tasks to be postponed.
func (m *Manager) postpone(pending *model.Task, list []model.Task) (found bool) {
	for i := range list {
		task := &list[i]
		if pending.ID == task.ID {
			continue
		}
		if pending.State != Running {
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
	job := r.job(&secret)
	err = r.client.Create(context.TODO(), &job)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.Started = &mark
	r.State = Running
	r.Job = path.Join(
		job.Namespace,
		job.Name)
	return
}

//
// Reflect finds the associated job and updates the task state.
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
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	mark := time.Now()
	status := job.Status
	for _, cnd := range status.Conditions {
		if cnd.Type == batch.JobFailed {
			r.State = Failed
			r.Terminated = &mark
			r.Error = "job failed."
			return
		}
		if status.Succeeded > 0 {
			r.State = Succeeded
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
		err = liberr.Wrap(err)
		return
	}

	return
}

//
// job build the Job.
func (r *Task) job(secret *core.Secret) (job batch.Job) {
	template := r.template(secret)
	backOff := int32(1)
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
		Name:            "main",
		Image:           r.Image,
		ImagePullPolicy: core.PullAlways,
		WorkingDir:      Settings.Addon.Path.WorkingDir,
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
		Encryption  struct {
			Passphrase string
		}
	}
	Addon interface{}
}