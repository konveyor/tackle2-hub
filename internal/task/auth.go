package task

import (
	"context"
	"fmt"
	"path"

	"github.com/golang-jwt/jwt/v5"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// Validator validates task tokens.
type Validator struct {
	// k8s client.
	Client k8s.Client
}

// Valid token when:
//   - The token references a task.
//   - The task is valid and running.
//   - The task pod valid and pending|running.
func (r *Validator) Valid(token *jwt.Token, db *gorm.DB) (err error) {
	claims := token.Claims.(jwt.MapClaims)
	v, found := claims["task"]
	id, cast := v.(float64)
	if !found || !cast {
		return
	}
	task := &model.Task{}
	err = db.First(task, id).Error
	if err != nil {
		err = &auth.NotValid{
			Token: token.Raw,
			Reason: fmt.Sprintf(
				"Task (%d) referenced by token: not found.",
				uint64(id)),
		}
		return
	}
	switch task.State {
	case Pending,
		Running:
	default:
		err = &auth.NotValid{
			Token: token.Raw,
			Reason: fmt.Sprintf(
				"Task (%d) referenced by token: not running.",
				uint64(id)),
		}
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
		err = &auth.NotValid{
			Token: token.Raw,
			Reason: fmt.Sprintf(
				"Pod (%s) referenced by token: not found.",
				pod.Name),
		}
		return
	}
	switch pod.Status.Phase {
	case core.PodPending,
		core.PodRunning:
	default:
		err = &auth.NotValid{
			Token: token.Raw,
			Reason: fmt.Sprintf(
				"Pod (%s) referenced by token: not pending|running. Phase detected: %s",
				task.Pod,
				pod.Status.Phase),
		}
		return
	}
	return
}
