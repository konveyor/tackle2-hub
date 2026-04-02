package binding

import (
	"errors"
	"os"
	"testing"

	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTaskGroup(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the task group to create
	taskGroup := &api.TaskGroup{
		Name: "Test Task Group",
		Data: api.Map{
			"mode": api.Map{
				"binary":   true,
				"withDeps": false,
			},
			"output": "/output/report",
		},
		State:    tasking.Created,
		Priority: 7,
	}

	// CREATE: Create the task group
	err := client.TaskGroup.Create(taskGroup)
	g.Expect(err).To(BeNil())
	g.Expect(taskGroup.ID).NotTo(BeZero())
	// Verify priority was adjusted (user priority must be >= 10)
	g.Expect(taskGroup.Priority).To(Equal(10))

	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroup.ID)
	})

	// LIST: List task groups and verify
	list, err := client.TaskGroup.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(BeNumerically(">", 0))
	found := false
	for _, tg := range list {
		if tg.ID == taskGroup.ID {
			found = true
			eq, report := cmp.Eq(taskGroup, &tg)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// GET: Retrieve the task group and verify it matches
	retrieved, err := client.TaskGroup.Get(taskGroup.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(taskGroup, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the task group
	taskGroup.Name = "Updated Test Task Group"
	taskGroup.Data = api.Map{
		"mode": api.Map{
			"binary":   false,
			"withDeps": true,
		},
		"output": "/output/report-updated",
	}

	err = client.TaskGroup.Update(taskGroup)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.TaskGroup.Get(taskGroup.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(taskGroup, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the task group
	err = client.TaskGroup.Delete(taskGroup.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.TaskGroup.Get(taskGroup.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTaskGroupPatch tests partial updates via Patch
func TestTaskGroupPatch(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a task group
	taskGroup := &api.TaskGroup{
		Name:  "Test Task Group for Patch",
		State: tasking.Created,
		Data: api.Map{
			"mode": api.Map{
				"binary": false,
			},
		},
	}
	err := client.TaskGroup.Create(taskGroup)
	g.Expect(err).To(BeNil())
	g.Expect(taskGroup.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroup.ID)
	})

	// PATCH: Partial update of the task group
	type TaskGroupPatch struct {
		Name string `json:"name"`
	}
	patch := &TaskGroupPatch{
		Name: "Patched Test Task Group",
	}

	err = client.TaskGroup.Patch(taskGroup.ID, patch)
	g.Expect(err).To(BeNil())

	// Update task group object to reflect the patch
	taskGroup.Name = "Patched Test Task Group"

	// GET: Retrieve again and verify patch
	patched, err := client.TaskGroup.Get(taskGroup.ID)
	g.Expect(err).To(BeNil())
	g.Expect(patched).NotTo(BeNil())
	g.Expect(patched.Name).To(Equal(patch.Name))
}

// TestTaskGroupSubmit tests submitting a task group
func TestTaskGroupSubmit(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a task group
	taskGroup := &api.TaskGroup{
		Name:  "Test Task Group for Submit",
		State: tasking.Created,
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
	}
	err := client.TaskGroup.Create(taskGroup)
	g.Expect(err).To(BeNil())
	g.Expect(taskGroup.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroup.ID)
	})

	// SUBMIT: Submit the task group
	err = client.TaskGroup.Submit(taskGroup.ID)
	g.Expect(err).To(BeNil())

	// GET: Retrieve and verify state changed
	submitted, err := client.TaskGroup.Get(taskGroup.ID)
	g.Expect(err).To(BeNil())
	g.Expect(submitted.State).NotTo(Equal(tasking.Created))
}

// TestTaskGroupPriorityAdjustment tests that task group priority is adjusted to be >= 10
func TestTaskGroupPriorityAdjustment(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a task group with priority < 10
	taskGroupLow := &api.TaskGroup{
		Name:     "Test Task Group Low Priority",
		State:    tasking.Created,
		Priority: 2,
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
	}
	err := client.TaskGroup.Create(taskGroupLow)
	g.Expect(err).To(BeNil())
	g.Expect(taskGroupLow.ID).NotTo(BeZero())
	// Verify priority was adjusted to 10
	g.Expect(taskGroupLow.Priority).To(Equal(10))

	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroupLow.ID)
	})

	// GET: Retrieve and verify adjusted priority persisted
	retrieved, err := client.TaskGroup.Get(taskGroupLow.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Priority).To(Equal(10))

	// CREATE: Create a task group with priority >= 10
	taskGroupHigh := &api.TaskGroup{
		Name:     "Test Task Group High Priority",
		State:    tasking.Created,
		Priority: 20,
		Data: api.Map{
			"mode": api.Map{
				"binary": false,
			},
		},
	}
	err = client.TaskGroup.Create(taskGroupHigh)
	g.Expect(err).To(BeNil())
	g.Expect(taskGroupHigh.ID).NotTo(BeZero())
	// Verify priority was NOT adjusted (stays 20)
	g.Expect(taskGroupHigh.Priority).To(Equal(20))

	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroupHigh.ID)
	})

	// GET: Retrieve and verify priority stayed at 20
	retrieved, err = client.TaskGroup.Get(taskGroupHigh.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Priority).To(Equal(20))

	// UPDATE: Update task group with priority < 10
	taskGroupLow.Priority = 8
	err = client.TaskGroup.Update(taskGroupLow)
	g.Expect(err).To(BeNil())
	// Verify priority was adjusted to 10
	g.Expect(taskGroupLow.Priority).To(Equal(10))

	// GET: Retrieve and verify adjusted priority after update
	updated, err := client.TaskGroup.Get(taskGroupLow.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated.Priority).To(Equal(10))
}

// TestTaskGroupBucket tests task group bucket file operations
func TestTaskGroupBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	// CREATE: Create a task group
	taskGroup := &api.TaskGroup{
		Name:  "Test Task Group for Bucket",
		State: tasking.Created,
		Data: api.Map{
			"mode": api.Map{
				"binary": false,
			},
		},
	}
	err := client.TaskGroup.Create(taskGroup)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.TaskGroup.Delete(taskGroup.ID)
	})

	// Get the task group bucket
	selected := client.TaskGroup.Select(taskGroup.ID)

	// PUT: Upload a file to the bucket
	tmpFile := "/tmp/test-taskgroup-bucket-source.txt"
	testContent := []byte("This is test content for the task group bucket")
	err = os.WriteFile(tmpFile, testContent, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile)

	err = selected.Bucket.Put(tmpFile, "test-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Download the file
	tmpDest := "/tmp/test-taskgroup-bucket-dest.txt"
	defer os.Remove(tmpDest)
	err = selected.Bucket.Get("test-file.txt", tmpDest)
	g.Expect(err).To(BeNil())
	content, err := os.ReadFile(tmpDest)
	g.Expect(err).To(BeNil())
	g.Expect(content).To(Equal(testContent))

	// PUT: Upload another file with nested path
	tmpFile2 := "/tmp/test-taskgroup-bucket-nested.txt"
	nestedContent := []byte("nested content")
	err = os.WriteFile(tmpFile2, nestedContent, 0644)
	g.Expect(err).To(BeNil())
	defer os.Remove(tmpFile2)

	err = selected.Bucket.Put(tmpFile2, "test-dir/nested-file.txt")
	g.Expect(err).To(BeNil())

	// GET: Download nested file
	tmpDest2 := "/tmp/test-taskgroup-bucket-nested-dest.txt"
	defer os.Remove(tmpDest2)
	err = selected.Bucket.Get("test-dir/nested-file.txt", tmpDest2)
	g.Expect(err).To(BeNil())
	content, err = os.ReadFile(tmpDest2)
	g.Expect(err).To(BeNil())
	g.Expect(content).To(Equal(nestedContent))

	// DELETE: Delete a file
	err = selected.Bucket.Delete("test-file.txt")
	g.Expect(err).To(BeNil())

	// Verify deletion
	err = selected.Bucket.Get("test-file.txt", tmpDest)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
