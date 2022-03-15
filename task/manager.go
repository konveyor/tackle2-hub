package task

import (
	"context"
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

const (
	ReaperUnit = time.Hour
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
		Log.Info("Reaper started.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(ReaperUnit)
				m.reapTasks()
				m.reapGroups()
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
		"status IN ?",
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
		switch ready.Status {
		case Ready,
			Postponed:
			if m.postpone(ready, list) {
				ready.Status = Postponed
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
// updateRunning tasks to reflect job status.
func (m *Manager) updateRunning() {
	list := []model.Task{}
	result := m.DB.Find(&list, "status", Running)
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
// reapTasks reaps tasks.
func (m *Manager) reapTasks() {
	list := []model.Task{}
	result := m.DB.Find(
		&list,
		"status IN ?",
		[]string{
			Succeeded,
			Failed,
		})
	Log.Trace(result.Error)
	if result.Error != nil {
		return
	}
	for i := range list {
		task := &list[i]
		if m.mayDelete(task) {
			result := m.DB.Delete(task)
			Log.Trace(result.Error)
			continue
		}
		if task.Purged {
			continue
		}
		if !m.mayPurge(task) {
			continue
		}
		task.Purged = true
		err := task.Purge()
		Log.Trace(err)
		if err != nil {
			continue
		}
		Log.Info("Task bucket purged.", "id", task.ID)
		result := m.DB.Save(task)
		Log.Trace(result.Error)
	}

	return
}

//
// reapGroups reaps groups.
func (m *Manager) reapGroups() (err error) {
	list := []model.TaskGroup{}
	db := m.DB.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		err = result.Error
		return
	}
	for i := range list {
		g := &list[i]
		if m.mayDeleteGroup(g) {
			result := m.DB.Delete(g)
			Log.Trace(result.Error)
			if result.Error == nil {
				Log.Info("Group deleted.", "id", g.ID)
			}
			continue
		}
		if g.Purged {
			continue
		}
		if !m.mayPurgeGroup(g) {
			continue
		}
		Log.Info("Group bucket purged.", "id", g.ID)
		g.Purged = true
		err := g.Purge()
		Log.Trace(err)
		if err != nil {
			continue
		}
		Log.Info("Group bucket purged.", "id", g.ID)
		result := m.DB.Save(g)
		Log.Trace(result.Error)
	}
	return
}

//
// mayPurge determines if a task (bucket) may be purged.
// May be purged when:
//   - Not associated with a group.
//   - Terminated for defined period.
func (m *Manager) mayPurge(task *model.Task) (may bool) {
	if task.TaskGroupID != nil {
		return
	}
	switch task.Status {
	case Succeeded:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
		may = time.Since(mark) > d
	case Failed:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Failed) * ReaperUnit
		may = time.Since(mark) > d
	}
	return
}

//
// mayDelete determines if a task may be deleted.
// May be deleted:
//   - Not associated with an application.
//   - Never submitted or terminated for defined period.
func (m *Manager) mayDelete(task *model.Task) (approved bool) {
	if task.ApplicationID != nil {
		return
	}
	switch task.Status {
	case Created:
		mark := task.CreateTime
		d := time.Duration(
			Settings.Hub.Task.Reaper.Created) * ReaperUnit
		approved = time.Since(mark) > d
	case Succeeded:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Succeeded) * ReaperUnit
		approved = time.Since(mark) > d
	case Failed:
		mark := *task.Terminated
		d := time.Duration(
			Settings.Hub.Task.Reaper.Failed) * ReaperUnit
		approved = time.Since(mark) > d
	}
	return
}

//
// mayDeleteGroup determines if a group may be deleted.
// May be deleted when:
//   - Empty for defined period.
func (m *Manager) mayDeleteGroup(g *model.TaskGroup) (approved bool) {
	empty := len(g.Tasks) == 0
	mark := g.CreateTime
	d := time.Duration(
		Settings.Hub.Task.Reaper.Created) * ReaperUnit
	approved = empty && time.Since(mark) > d
	return
}

//
// mayPurgeGroup determines if a group may be purged.
// May be purged when:
//   - All tasks may purge.
func (m *Manager) mayPurgeGroup(g *model.TaskGroup) (approved bool) {
	nMayPurge := 0
	for i := range g.Tasks {
		task := &g.Tasks[i]
		task.TaskGroupID = nil
		if m.mayPurge(task) {
			nMayPurge++
		}
	}
	approved = nMayPurge == len(g.Tasks)
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
		} else {
			err = liberr.Wrap(err)
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
		err = liberr.Wrap(err)
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
