package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestRuleSet(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the ruleset to create
	ruleSet := &api.RuleSet{
		Name:        "Test RuleSet",
		Description: "Test ruleset description",
		Rules:       []api.Rule{},
	}

	// Get seeded.
	seeded, err := client.RuleSet.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the ruleset
	err = client.RuleSet.Create(ruleSet)
	g.Expect(err).To(BeNil())
	g.Expect(ruleSet.ID).NotTo(BeZero())

	defer func() {
		_ = client.RuleSet.Delete(ruleSet.ID)
	}()

	// GET: List rulesets
	list, err := client.RuleSet.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(ruleSet, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the ruleset and verify it matches
	retrieved, err := client.RuleSet.Get(ruleSet.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(ruleSet, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the ruleset
	ruleSet.Name = "Updated Test RuleSet"
	ruleSet.Description = "Updated ruleset description"

	err = client.RuleSet.Update(ruleSet)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.RuleSet.Get(ruleSet.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(ruleSet, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the ruleset
	err = client.RuleSet.Delete(ruleSet.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.RuleSet.Get(ruleSet.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
