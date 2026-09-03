package reaper

import (
	"context"
	"errors"
	"strconv"
	"time"

	agent "github.com/konveyor/agentic-controller/api/v1alpha1"
	agenticmeta "github.com/konveyor/tackle2-hub/internal/agentic"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"gorm.io/gorm"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const tokenSecretOrphanGrace = 5 * time.Minute

// AgenticTokenReaper revokes Hub API tokens whose owning run has finished or
// disappeared. The Secret finalizer keeps the token ID available when
// Kubernetes garbage collection races the reaper.
type AgenticTokenReaper struct {
	Client      k8s.Client
	RevokeToken func(uint) error
}

// Run turns the agentic token janitor once.
func (r *AgenticTokenReaper) Run() {
	if r.Client == nil {
		return
	}
	ctx := context.Background()
	list := &core.SecretList{}
	err := r.Client.List(
		ctx,
		list,
		k8s.InNamespace(Settings.Hub.Namespace),
		k8s.MatchingLabels{agenticmeta.TokenSecretLabel: "true"})
	if err != nil {
		Log.Error(err, "Failed to list agentic token Secrets.")
		return
	}
	for i := range list.Items {
		secret := &list.Items[i]
		reap, err := r.shouldRevoke(ctx, secret)
		if err != nil {
			Log.Error(err, "Failed to inspect agentic token owner.", "secret", secret.Name)
			continue
		}
		if reap {
			r.revoke(ctx, secret)
		}
	}
}

func (r *AgenticTokenReaper) shouldRevoke(ctx context.Context, secret *core.Secret) (bool, error) {
	if !secret.DeletionTimestamp.IsZero() {
		return true, nil
	}
	owner := runOwner(secret.OwnerReferences)
	if owner == nil {
		return secret.CreationTimestamp.Add(tokenSecretOrphanGrace).Before(time.Now()), nil
	}
	key := types.NamespacedName{Namespace: secret.Namespace, Name: owner.Name}
	switch owner.Kind {
	case "AgentRun":
		run := &agent.AgentRun{}
		if err := r.Client.Get(ctx, key, run); err != nil {
			return r.missingOwner(secret, err)
		}
		if owner.UID != "" && owner.UID != run.UID {
			return true, nil
		}
		return terminal(run.Status.Phase) || !run.DeletionTimestamp.IsZero(), nil
	case "AgentWorkflowRun":
		run := &agent.AgentWorkflowRun{}
		if err := r.Client.Get(ctx, key, run); err != nil {
			return r.missingOwner(secret, err)
		}
		if owner.UID != "" && owner.UID != run.UID {
			return true, nil
		}
		return terminal(run.Status.Phase) || !run.DeletionTimestamp.IsZero(), nil
	default:
		return false, nil
	}
}

func (r *AgenticTokenReaper) missingOwner(secret *core.Secret, err error) (bool, error) {
	if !k8serr.IsNotFound(err) {
		return false, err
	}
	// Allow informer caches to observe a just-created owner before treating
	// the Secret as orphaned.
	return secret.CreationTimestamp.Add(tokenSecretOrphanGrace).Before(time.Now()), nil
}

func (r *AgenticTokenReaper) revoke(ctx context.Context, secret *core.Secret) {
	rawID, found := secret.Data[agenticmeta.TokenIDKey]
	if !found {
		Log.Info("Agentic token Secret has no token ID.", "secret", secret.Name)
		r.release(ctx, secret)
		return
	}
	tokenID, err := strconv.ParseUint(string(rawID), 10, 64)
	if err != nil {
		Log.Error(err, "Agentic token Secret has an invalid token ID.", "secret", secret.Name)
		r.release(ctx, secret)
		return
	}
	revoke := r.RevokeToken
	if revoke == nil {
		revoke = auth.Idp().Revoke
	}
	if err = revoke(uint(tokenID)); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		Log.Error(err, "Failed to revoke agentic token.", "id", tokenID)
		return
	}
	if err == nil {
		Log.Info("Agentic token revoked.", "id", tokenID, "secret", secret.Name)
	}
	r.release(ctx, secret)
}

func (r *AgenticTokenReaper) release(ctx context.Context, secret *core.Secret) {
	controllerutil.RemoveFinalizer(secret, agenticmeta.TokenSecretFinalizer)
	delete(secret.Labels, agenticmeta.TokenSecretLabel)
	if err := r.Client.Update(ctx, secret); err != nil && !k8serr.IsNotFound(err) {
		Log.Error(err, "Failed to release agentic token Secret.", "secret", secret.Name)
	}
}

func runOwner(refs []metav1.OwnerReference) *metav1.OwnerReference {
	for i := range refs {
		switch refs[i].Kind {
		case "AgentRun", "AgentWorkflowRun":
			return &refs[i]
		}
	}
	return nil
}

func terminal(phase agent.AgentRunPhase) bool {
	return phase == agent.AgentRunPhaseSucceeded || phase == agent.AgentRunPhaseFailed
}
