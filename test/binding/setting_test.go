package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestSetting(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the setting to create
	setting := &api.Setting{
		Key:   "test.setting.1",
		Value: "test-value-123",
	}

	// CREATE: Create the setting
	err := client.Setting.Create(setting)
	g.Expect(err).To(BeNil())

	t.Cleanup(func() {
		_ = client.Setting.Delete(setting.Key)
	})

	// GET: Retrieve the setting and verify it matches
	var retrievedValue string
	err = client.Setting.Get(setting.Key, &retrievedValue)
	g.Expect(err).To(BeNil())
	g.Expect(retrievedValue).To(Equal(setting.Value))

	// UPDATE: Modify the setting
	setting.Value = "updated-value-456"

	err = client.Setting.Update(setting)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	var updatedValue string
	err = client.Setting.Get(setting.Key, &updatedValue)
	g.Expect(err).To(BeNil())
	g.Expect(updatedValue).To(Equal(setting.Value))

	// DELETE: Remove the setting
	err = client.Setting.Delete(setting.Key)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	var deletedValue string
	err = client.Setting.Get(setting.Key, &deletedValue)
	g.Expect(err).NotTo(BeNil())
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
