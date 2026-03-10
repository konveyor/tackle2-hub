package task

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	internaltask "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestScheduler tests the task scheduler with simulator.
func TestScheduler(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup test environment
	ctx := setup(g)
	defer ctx.teardown()

	// Seed database with common test data
	app := ctx.seed(g)

	// Create 3 tasks with kind=analyzer, state=Ready
	for i := 1; i <= 3; i++ {
		task := &model.Task{
			Name:          "test-task-" + strconv.Itoa(i),
			Kind:          "analyzer",
			State:         internaltask.Ready,
			ApplicationID: &app.ID,
		}
		err := ctx.DB.Create(task).Error
		g.Expect(err).To(gomega.BeNil())
		g.Expect(task.ID).To(gomega.Equal(uint(i)))
	}

	// Create and start the manager
	ctx.newManager(g)

	// List pods to verify they were created
	podList := &core.PodList{}
	err := ctx.Client.List(context.Background(), podList, &client.ListOptions{
		Namespace: settings.Settings.Hub.Namespace,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(podList.Items)).To(gomega.BeNumerically(">=", 1))

	// Check the first created pod for expected configuration
	pod := &podList.Items[0]
	g.Expect(pod.Labels).To(gomega.HaveKey(internaltask.TaskLabel))
	g.Expect(pod.Labels[internaltask.AppLabel]).To(gomega.Equal("tackle"))
	g.Expect(pod.Labels[internaltask.RoleLabel]).To(gomega.Equal("task"))

	// Verify pod has 2 containers: addon (analyzer) and java extension
	g.Expect(pod.Spec.Containers).To(gomega.HaveLen(2))

	// First container should be the addon container
	addonContainer := pod.Spec.Containers[0]
	g.Expect(addonContainer.Name).To(gomega.Equal("addon"))
	g.Expect(addonContainer.Image).To(gomega.Equal("quay.io/konveyor/tackle2-addon-analyzer"))

	// Second container should be the java extension
	javaContainer := pod.Spec.Containers[1]
	g.Expect(javaContainer.Name).To(gomega.Equal("java"))
	g.Expect(javaContainer.Image).To(gomega.Equal("quay.io/konveyor/java-external-provider:latest"))

	// Verify java extension has expected environment variables from extension spec
	javaEnvVars := make(map[string]string)
	for _, env := range javaContainer.Env {
		if env.Value != "" {
			javaEnvVars[env.Name] = env.Value
		}
	}
	g.Expect(javaEnvVars).To(gomega.HaveKey("PORT"))
	g.Expect(javaEnvVars["PORT"]).To(gomega.Equal("8000"))
	g.Expect(javaEnvVars).To(gomega.HaveKey("MAVEN_OPTS"))
	g.Expect(javaEnvVars["MAVEN_OPTS"]).To(gomega.Equal("-Dmaven.repo.local=/cache/m2"))
	g.Expect(javaEnvVars).To(gomega.HaveKey("JAVA_TOOL_OPTIONS"))

	// Verify addon container environment variables
	envVars := make(map[string]string)
	for _, env := range addonContainer.Env {
		if env.Value != "" {
			envVars[env.Name] = env.Value
		}
	}

	g.Expect(envVars).To(gomega.HaveKey(settings.EnvAddonHomeDir))
	g.Expect(envVars[settings.EnvAddonHomeDir]).To(gomega.Equal(settings.Settings.Addon.HomeDir))

	g.Expect(envVars).To(gomega.HaveKey(settings.EnvSharedDir))
	g.Expect(envVars[settings.EnvSharedDir]).To(gomega.Equal(settings.Settings.Addon.SharedDir))

	g.Expect(envVars).To(gomega.HaveKey(settings.EnvCacheDir))
	g.Expect(envVars[settings.EnvCacheDir]).To(gomega.Equal(settings.Settings.Addon.CacheDir))

	g.Expect(envVars).To(gomega.HaveKey(settings.EnvHubBaseURL))
	g.Expect(envVars[settings.EnvHubBaseURL]).To(gomega.Equal(settings.Settings.Addon.Hub.URL))

	g.Expect(envVars).To(gomega.HaveKey(settings.EnvTask))

	// Verify addon container has propagated environment variables from java extension
	// Extension env vars are prefixed with _EXT_<EXTENSION-NAME>_<VAR-NAME> (uppercased)
	g.Expect(envVars).To(gomega.HaveKey("_EXT_JAVA_PORT"))
	g.Expect(envVars).To(gomega.HaveKey("_EXT_JAVA_MAVEN_OPTS"))
	g.Expect(envVars["_EXT_JAVA_MAVEN_OPTS"]).To(gomega.Equal("-Dmaven.repo.local=/cache/m2"))
	g.Expect(envVars).To(gomega.HaveKey("_EXT_JAVA_JAVA_TOOL_OPTIONS"))

	// Verify EnvHubToken is set from secret
	hasTokenEnv := false
	for _, env := range addonContainer.Env {
		if env.Name == settings.EnvHubToken {
			hasTokenEnv = true
			g.Expect(env.ValueFrom).ToNot(gomega.BeNil())
			g.Expect(env.ValueFrom.SecretKeyRef).ToNot(gomega.BeNil())
			g.Expect(env.ValueFrom.SecretKeyRef.Key).To(gomega.Equal(settings.EnvHubToken))
			break
		}
	}
	g.Expect(hasTokenEnv).To(gomega.BeTrue())

	// Verify expected volume mounts
	mountPaths := make(map[string]string)
	for _, mount := range addonContainer.VolumeMounts {
		mountPaths[mount.Name] = mount.MountPath
	}

	g.Expect(mountPaths).To(gomega.HaveKey(internaltask.Addon))
	g.Expect(mountPaths[internaltask.Addon]).To(gomega.Equal(settings.Settings.Addon.HomeDir))

	g.Expect(mountPaths).To(gomega.HaveKey(internaltask.Shared))
	g.Expect(mountPaths[internaltask.Shared]).To(gomega.Equal(settings.Settings.Addon.SharedDir))

	g.Expect(mountPaths).To(gomega.HaveKey(internaltask.Cache))
	g.Expect(mountPaths[internaltask.Cache]).To(gomega.Equal(settings.Settings.Addon.CacheDir))

	// Wait for pods to progress through lifecycle (Pending -> Running -> Succeeded)
	// With settings.Settings.Frequency.Task at 100ms and pod transitions at 1s each,
	// we need ~2.5s for all 3 tasks to complete (with RuleUnique postponing task 3)
	time.Sleep(2500 * time.Millisecond)

	// Retrieve updated pod list to verify they progressed through lifecycle
	err = ctx.Client.List(context.Background(), podList, &client.ListOptions{
		Namespace: settings.Settings.Hub.Namespace,
	})
	g.Expect(err).To(gomega.BeNil())

	// Verify pods progressed through lifecycle
	succeededPods := 0
	runningOrCompletedPods := 0
	for _, p := range podList.Items {
		if p.Status.Phase == core.PodSucceeded {
			succeededPods++
			// Verify container terminated successfully
			g.Expect(p.Status.ContainerStatuses).ToNot(gomega.BeEmpty())
			for _, status := range p.Status.ContainerStatuses {
				g.Expect(status.State.Terminated).ToNot(gomega.BeNil())
				g.Expect(status.State.Terminated.ExitCode).To(gomega.Equal(int32(0)))
			}
		}
		if p.Status.Phase == core.PodRunning || p.Status.Phase == core.PodSucceeded {
			runningOrCompletedPods++
		}
	}
	// Verify at least one pod reached running or succeeded state
	g.Expect(runningOrCompletedPods).To(gomega.BeNumerically(">=", 1))

	// Verify tasks in database were updated to Succeeded state
	// Note: Due to the RuleUnique policy, only 2 tasks can run concurrently
	// for the same application/kind combination. The third task will be postponed
	// until one of the first two completes.
	var tasks []*model.Task
	err = ctx.DB.Find(&tasks, "state IN ?", []string{
		internaltask.Succeeded,
		internaltask.Running,
		internaltask.Pending,
	}).Error
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(tasks)).To(gomega.BeNumerically(">=", 2))

	// Find succeeded tasks
	var succeededTasks []*model.Task
	for _, task := range tasks {
		if task.State == internaltask.Succeeded {
			succeededTasks = append(succeededTasks, task)
		}
	}
	g.Expect(len(succeededTasks)).To(gomega.BeNumerically(">=", 1))

	// Verify at least one task has proper state transitions recorded
	task := succeededTasks[0]
	g.Expect(task.State).To(gomega.Equal(internaltask.Succeeded))
	g.Expect(task.Started).ToNot(gomega.BeNil())
	g.Expect(task.Terminated).ToNot(gomega.BeNil())
	g.Expect(task.Pod).ToNot(gomega.BeEmpty())
}
