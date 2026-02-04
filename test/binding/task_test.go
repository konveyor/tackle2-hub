package binding

import (
	"errors"
	"testing"

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
		State:    "Created",
		Priority: 5,
	}

	// CREATE: Create the task
	err := client.Task.Create(task)
	g.Expect(err).To(BeNil())
	g.Expect(task.ID).NotTo(BeZero())

	defer func() {
		_ = client.Task.Delete(task.ID)
	}()

	// GET: List tasks
	list, err := client.Task.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list) > 0).To(BeTrue())
	found := false
	for _, item := range list {
		if item.ID == task.ID {
			found = true
			eq, report := cmp.Eq(task, &item)
			g.Expect(eq).To(BeTrue(), report)
			break
		}
	}
	g.Expect(found).To(BeTrue())

	// GET: Retrieve the task and verify it matches
	retrieved, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(task, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the task
	task.Name = "Updated Test Task"
	task.State = "Running"
	task.Priority = 10

	err = client.Task.Update(task)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Name).To(Equal("Updated Test Task"))
	g.Expect(updated.State).To(Equal("Running"))
	g.Expect(updated.Priority).To(Equal(10))

	// PATCH: Test partial update
	type TaskPatch struct {
		Name string `json:"name"`
	}
	patch := &TaskPatch{
		Name: "Patched Test Task",
	}

	err = client.Task.Patch(task.ID, patch)
	g.Expect(err).To(BeNil())

	// GET: Verify patch
	patched, err := client.Task.Get(task.ID)
	g.Expect(err).To(BeNil())
	g.Expect(patched).NotTo(BeNil())
	g.Expect(patched.Name).To(Equal("Patched Test Task"))
	g.Expect(patched.State).To(Equal("Running")) // Should remain unchanged

	// DELETE: Remove the task
	err = client.Task.Delete(task.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Task.Get(task.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
