package binding

import (
	"errors"
	"testing"
	"time"

	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTask(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the task to create
	task := &api.Task{
		Name:  "Test Task",
		Addon: "analyzer",
		Kind:  "test-kind",
		Data: api.Map{
			"key1": "value1",
			"key2": "value2",
		},
		State:    tasking.Created,
		Priority: 5,
	}

	// CREATE: Create the task
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Task.Delete(task.ID)
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
	task.Priority = 10

	err = client.Task.Update(task)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	time.Sleep(time.Second)
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
	time.Sleep(time.Second)
	patched, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(patched).NotTo(BeNil())
	eq, report = cmp.Eq(task, patched, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the task
	err = client.Task.Delete(task.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	time.Sleep(time.Second)
	_, err = client.Task.Get(task.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
