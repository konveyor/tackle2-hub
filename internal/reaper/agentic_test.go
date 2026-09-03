package reaper

import (
	"context"
	"fmt"
	"testing"
	"time"

	agent "github.com/konveyor/agentic-controller/api/v1alpha1"
	agenticmeta "github.com/konveyor/tackle2-hub/internal/agentic"
	"github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAgenticTokenReaperRevokesTerminalRunToken(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	const tokenID = 41
	namespace := Settings.Hub.Namespace
	uid := types.UID("run-uid")
	run := &agent.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "failed-run", Namespace: namespace, UID: uid},
		Status:     agent.AgentRunStatus{Phase: agent.AgentRunPhaseFailed},
	}
	secret := agenticTokenSecret("failed-run-token", namespace, tokenID)
	secret.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: agent.GroupVersion.String(),
		Kind:       "AgentRun",
		Name:       run.Name,
		UID:        uid,
	}}
	client := fake.NewClientBuilder().WithScheme(agenticScheme(t)).WithObjects(run, secret).Build()

	var revoked []uint
	(&AgenticTokenReaper{
		Client: client,
		RevokeToken: func(id uint) error {
			revoked = append(revoked, id)
			return nil
		},
	}).Run()

	g.Expect(revoked).To(gomega.Equal([]uint{tokenID}))
	fetched := &core.Secret{}
	g.Expect(client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      secret.Name,
	}, fetched)).To(gomega.Succeed())
	g.Expect(fetched.Finalizers).ToNot(gomega.ContainElement(agenticmeta.TokenSecretFinalizer))
	g.Expect(fetched.Labels).ToNot(gomega.HaveKey(agenticmeta.TokenSecretLabel))
}

func TestAgenticTokenReaperKeepsActiveRunToken(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	const tokenID = 42
	namespace := Settings.Hub.Namespace
	uid := types.UID("run-uid")
	run := &agent.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "active-run", Namespace: namespace, UID: uid},
		Status:     agent.AgentRunStatus{Phase: agent.AgentRunPhaseRunning},
	}
	secret := agenticTokenSecret("active-run-token", namespace, tokenID)
	secret.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: agent.GroupVersion.String(),
		Kind:       "AgentRun",
		Name:       run.Name,
		UID:        uid,
	}}
	client := fake.NewClientBuilder().WithScheme(agenticScheme(t)).WithObjects(run, secret).Build()

	var revoked []uint
	(&AgenticTokenReaper{
		Client: client,
		RevokeToken: func(id uint) error {
			revoked = append(revoked, id)
			return nil
		},
	}).Run()

	g.Expect(revoked).To(gomega.BeEmpty())
	fetched := &core.Secret{}
	g.Expect(client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      secret.Name,
	}, fetched)).To(gomega.Succeed())
	g.Expect(fetched.Finalizers).To(gomega.ContainElement(agenticmeta.TokenSecretFinalizer))
	g.Expect(fetched.Labels).To(gomega.HaveKeyWithValue(agenticmeta.TokenSecretLabel, "true"))
}

func TestAgenticTokenReaperRevokesOrphanAfterGrace(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	const tokenID = 43
	secret := agenticTokenSecret("orphan-token", Settings.Hub.Namespace, tokenID)
	secret.CreationTimestamp = metav1.NewTime(time.Now().Add(-tokenSecretOrphanGrace - time.Minute))
	client := fake.NewClientBuilder().WithScheme(agenticScheme(t)).WithObjects(secret).Build()

	var revoked []uint
	(&AgenticTokenReaper{
		Client: client,
		RevokeToken: func(id uint) error {
			revoked = append(revoked, id)
			return nil
		},
	}).Run()

	g.Expect(revoked).To(gomega.Equal([]uint{tokenID}))
}

func agenticScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := core.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := agent.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

func agenticTokenSecret(name, namespace string, tokenID uint) *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Labels:     map[string]string{agenticmeta.TokenSecretLabel: "true"},
			Finalizers: []string{agenticmeta.TokenSecretFinalizer},
		},
		Data: map[string][]byte{agenticmeta.TokenIDKey: []byte(fmt.Sprint(tokenID))},
	}
}
