package binding

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/filter"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTaskWithApplication(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the task to reference
	app := &api.Application{
		Name:        "Test App for Task",
		Description: "Application for task testing",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Define the task to create with complex Data and Application reference
	task := &api.Task{
		Name:       "Test Task",
		Kind:       "analyzer",
		Addon:      "analyzer",
		Extensions: []string{"java", "nodejs"},
		Locator:    "task-locator-123",
		Policy: api.TaskPolicy{
			Isolated: true,
		},
		TTL: api.TTL{
			Created:   30,
			Pending:   60,
			Running:   120,
			Succeeded: 180,
			Failed:    240,
		},
		Data: api.Map{
			"mode": api.Map{
				"binary":       true,
				"withDeps":     false,
				"artifact":     "",
				"diva":         true,
				"csv":          false,
				"dependencies": true,
			},
			"output":  "/windup/report",
			"rules":   []string{"ruleA", "ruleB"},
			"targets": []string{"cloud-readiness"},
			"scope": api.Map{
				"packages": api.Map{
					"included": []string{"com.example"},
					"excluded": []string{"com.example.test"},
				},
			},
		},
		Application: &api.Ref{
			ID:   app.ID,
			Name: app.Name,
		},
		State:    tasking.Created,
		Priority: 5,
	}

	// CREATE: Create the task
	err = client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())

	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// GET: List tasks
	list, err := client.Task.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(task, &list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the task and verify it matches
	retrieved, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(task, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the task
	task.Name = "Updated Test Task"
	task.Locator = "updated-locator-456"
	task.Priority = 10
	task.Policy = api.TaskPolicy{
		Isolated: false,
	}
	task.TTL = api.TTL{
		Created:   60,
		Pending:   90,
		Running:   180,
		Succeeded: 240,
		Failed:    300,
	}
	task.Data = api.Map{
		"mode": api.Map{
			"binary":       false,
			"withDeps":     true,
			"artifact":     "app.jar",
			"diva":         false,
			"csv":          true,
			"dependencies": false,
		},
		"output":  "/windup/report-updated",
		"rules":   []string{"ruleC", "ruleD", "ruleE"},
		"targets": []string{"quarkus", "cloud-readiness"},
		"scope": api.Map{
			"packages": api.Map{
				"included": []string{"com.example", "com.test"},
				"excluded": []string{},
			},
		},
	}

	err = client.Task.Update(task)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(task, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// PATCH: Partial update of the task
	type TaskPatch struct {
		Name string `json:"name"`
	}
	patch := &TaskPatch{
		Name: "Patched Test Task",
	}

	err = client.Task.Patch(task.ID, patch)
	g.Expect(err).To(BeNil())

	// Update task object to reflect the patch
	task.Name = "Patched Test Task"

	// GET: Retrieve again and verify patch
	patched, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(patched).NotTo(BeNil())
	eq, report = cmp.Eq(task, patched, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the task
	ctx, cfn := context.WithTimeout(
		context.Background(),
		time.Minute)
	defer cfn()
	err = client.Task.Select(task.ID).Blocking.Delete(ctx)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Task.Get(task.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTaskWithPlatform tests creating a task with Platform reference instead of Application
func TestTaskWithPlatform(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a platform for the task to reference
	platform := &api.Platform{
		Kind: "kubernetes",
		Name: "Test Platform",
		URL:  "https://test-platform.example.com",
	}
	err := client.Platform.Create(platform)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Platform.Delete(platform.ID)
	})

	// CREATE: Create a task with Platform reference
	task := &api.Task{
		Name: "Test Task with Platform",
		Kind: "application-manifest",
		Data: api.Map{
			"source": "/source/path",
			"target": "kubernetes",
			"output": "/output/path",
		},
		Platform: &api.Ref{
			ID:   platform.ID,
			Name: platform.Name,
		},
		State:    tasking.Created,
		Priority: 3,
	}

	err = client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// GET: Retrieve the task and verify it matches
	retrieved, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.Platform).NotTo(BeNil())
	g.Expect(retrieved.Platform.ID).To(Equal(platform.ID))
	eq, report := cmp.Eq(task, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the task
	ctx, cfn := context.WithTimeout(
		context.Background(),
		time.Minute)
	defer cfn()
	err = client.Task.Select(task.ID).Blocking.Delete(ctx)
	g.Expect(err).To(BeNil())

	// Verify deletion
	_, err = client.Task.Get(task.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTaskBulkCancel tests canceling multiple tasks using filter
func TestTaskCancel(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create task
	task := &api.Task{
		Name:     "Test Task 1",
		Addon:    "analyzer",
		State:    tasking.Created,
		Priority: 5,
	}
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// CANCEL: task
	ctx, cfn := context.WithTimeout(
		context.Background(),
		time.Minute)
	defer cfn()
	err = client.Task.Select(task.ID).Blocking.Cancel(ctx)
	g.Expect(err).To(BeNil())

	// Verify task was canceled
	retrieved, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.State).To(Equal(tasking.Canceled))
}

// TestTaskBulkCancel tests canceling multiple tasks using filter
func TestTaskBulkCancel(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create first task
	task1 := &api.Task{
		Name:     "Test Task 1",
		Addon:    "analyzer",
		State:    tasking.Created,
		Priority: 5,
	}
	err := client.Task.Create(task1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task1.ID).Blocking.Delete(ctx)
	})

	// Create second task
	task2 := &api.Task{
		Name:     "Test Task 2",
		Addon:    "analyzer",
		State:    tasking.Created,
		Priority: 5,
	}
	err = client.Task.Create(task2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task2.ID).Blocking.Delete(ctx)
	})

	// Create third task
	task3 := &api.Task{
		Name:     "Test Task 3",
		Addon:    "analyzer",
		State:    tasking.Created,
		Priority: 5,
	}
	err = client.Task.Create(task3)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task3.ID).Blocking.Delete(ctx)
	})

	// BULK CANCEL: Cancel tasks using filter
	f := binding.Filter{}
	f.And("id").Eq(filter.Any{task1.ID, task2.ID, task3.ID})
	err = client.Task.BulkCancel(f)
	g.Expect(err).To(BeNil())

	// Verify tasks were canceled
	canceled := []uint{task1.ID, task2.ID, task3.ID}
	isDone := func() (done bool) {
		for _, id := range canceled {
			var task *api.Task
			task, err = client.Task.Get(id)
			if err != nil {
				return
			}
			if task.State != tasking.Canceled {
				return
			}
		}
		done = true
		return
	}
	g.Eventually(isDone, 30*time.Second, time.Second).
		Should(BeTrue(), "Tasks should have been canceled")
}

// TestTaskBucket tests task bucket file operations
func TestTaskBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a task
	task := &api.Task{
		Name:  "Test Task for Bucket",
		Kind:  "analyzer",
		State: tasking.Created,
	}
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// Get the task bucket
	bucket := client.Task.Bucket(task.ID)

	// PUT: Upload a file to the bucket
	tmpFile := "/tmp/test-task-bucket-source.txt"
	testContent := []byte("This is test content for the task bucket")
	err = os.WriteFile(tmpFile, testContent, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile)

	err = bucket.Put(tmpFile, "test-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Download the file
	tmpDest := "/tmp/test-task-bucket-dest.txt"
	defer os.Remove(tmpDest)
	err = bucket.Get("test-file.txt", tmpDest)
	g.Expect(err).To(BeNil())
	content, err := os.ReadFile(tmpDest)
	g.Expect(err).To(BeNil())
	g.Expect(content).To(Equal(testContent))

	// DELETE: Delete a file
	err = bucket.Delete("test-file.txt")
	g.Expect(err).To(BeNil())

	// Verify deletion
	err = bucket.Get("test-file.txt", tmpDest)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTaskGetAttached tests retrieving a task with attached resources
func TestTaskGetAttached(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an application for the task to reference
	app := &api.Application{
		Name:        "Test App for Attached",
		Description: "Application for testing GetAttached",
	}
	err := client.Application.Create(app)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app.ID)
	})

	// Create first attached file
	tmpFile1 := "/tmp/test-attached-file1.txt"
	file1Content := []byte("This is attached file 1 content")
	err = os.WriteFile(tmpFile1, file1Content, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile1)

	attachedFile1, err := client.File.Put(tmpFile1)
	g.Expect(err).To(BeNil())
	g.Expect(attachedFile1).NotTo(BeNil())
	g.Expect(attachedFile1.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.File.Delete(attachedFile1.ID)
	})

	// Create second attached file
	tmpFile2 := "/tmp/test-attached-file2.txt"
	file2Content := []byte("This is attached file 2 content")
	err = os.WriteFile(tmpFile2, file2Content, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile2)

	attachedFile2, err := client.File.Put(tmpFile2)
	g.Expect(err).To(BeNil())
	g.Expect(attachedFile2).NotTo(BeNil())
	g.Expect(attachedFile2.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.File.Delete(attachedFile2.ID)
	})

	// CREATE: Create a task with Application reference and attached files
	task := &api.Task{
		Name: "Test Task for GetAttached",
		Kind: "analyzer",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
			"output": "/output/path",
		},
		Application: &api.Ref{
			ID:   app.ID,
			Name: app.Name,
		},
		Attached: []api.Attachment{
			{
				ID:   attachedFile1.ID,
				Name: "test-attached-file1.txt",
			},
			{
				ID:   attachedFile2.ID,
				Name: "test-attached-file2.txt",
			},
		},
		State:    tasking.Created,
		Priority: 5,
	}

	err = client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// GET ATTACHED: Download the task's attached resources as a tarball
	tmpDest := "/tmp/test-task-attached.tar.gz"
	defer os.Remove(tmpDest)
	err = client.Task.GetAttached(task.ID, tmpDest)
	g.Expect(err).To(BeNil())

	// Verify the tarball was created
	info, err := os.Stat(tmpDest)
	g.Expect(err).To(BeNil())
	g.Expect(info.Size()).To(BeNumerically(">", 0))
}

// TestTaskReport tests task report operations
func TestTaskReport(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a task
	task := &api.Task{
		Name:  "Test Task for Report",
		Addon: "analyzer",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
		State:    tasking.Created,
		Priority: 5,
	}
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())
	t.Cleanup(func() {
		ctx, cfn := context.WithTimeout(
			context.Background(),
			time.Minute)
		defer cfn()
		_ = client.Task.Select(task.ID).Blocking.Delete(ctx)
	})

	// Get selected task API
	selected := client.Task.Select(task.ID)

	// CREATE REPORT: Create a task report
	report := &api.TaskReport{
		Status:    "Completed",
		Total:     100,
		Completed: 100,
		Activity:  []string{"Step 1", "Step 2", "Step 3"},
		Result: api.Map{
			"findings": 42,
			"issues":   5,
		},
	}
	err = selected.Report.Create(report)
	g.Expect(err).To(BeNil())
	g.Expect(report.ID).NotTo(BeZero())

	// UPDATE REPORT: Update the task report
	report.Status = "Failed"
	report.Completed = 50
	report.Errors = []api.TaskError{
		{
			Severity:    "ERROR",
			Description: "Test error occurred",
		},
	}
	err = selected.Report.Update(report)
	g.Expect(err).To(BeNil())

	// GET: Retrieve the task and verify it has a report
	retrieved, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	// Note: We can't directly verify the report contents via Get,
	// but we can verify the task was retrieved successfully

	// DELETE REPORT: Delete the task report
	err = selected.Report.Delete()
	g.Expect(err).To(BeNil())
}
