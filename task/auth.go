package task

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	"path"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

//
// Validator validates task tokens.
type Validator struct {
	// k8s client.
	Client k8s.Client
	// DB client.
	DB *gorm.DB
}

//
// Valid token when:
//  - The token references a task.
//  - The task is valid and running.
//  - The task pod valid and pending|running.
func (r *Validator) Valid(token *jwt.Token) (valid bool) {
	var err error
	claims := token.Claims.(jwt.MapClaims)
	v, found := claims["task"]
	id, cast := v.(float64)
	if !found || !cast {
		Log.Info("Task not referenced by token.")
		return
	}
	task := &model.Task{}
	err = r.DB.First(task, id).Error
	if err != nil {
		Log.Info("Task referenced by token: not found.")
		return
	}
	switch task.State {
	case Pending,
		Running:
	default:
		Log.Info("Task referenced by token: not running.")
		return
	}
	pod := &core.Pod{}
	err = r.Client.Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: path.Dir(task.Pod),
			Name:      path.Base(task.Pod),
		},
		pod)
	if err != nil {
		Log.Info(
			"Pod referenced by token: not found.",
			"name",
			task.Pod)
		return
	}
	switch pod.Status.Phase {
	case core.PodPending,
		core.PodRunning:
	default:
		Log.Info(
			"Pod referenced by token: not running.",
			"name",
			task.Pod,
			"phase",
			pod.Status.Phase)
		return
	}
	valid = true
	return
}
