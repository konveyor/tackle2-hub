package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

func TestAnalysisProfile(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an identity for the profile to reference
	identity := &api.Identity{
		Name: "test-identity",
		Kind: "Test",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	defer func() {
		_ = client.Identity.Delete(identity.ID)
	}()

	// Define the profile to create
	profile := &api.AnalysisProfile{
		Name:        "Test Profile",
		Description: "This is a test analysis profile",
		Mode: api.ApMode{
			WithDeps: true,
		},
		Scope: api.ApScope{
			WithKnownLibs: true,
			Packages: api.InExList{
				Included: []string{"com.example.pkg1", "com.example.pkg2"},
				Excluded: []string{"com.example.pkg3"},
			},
		},
		Rules: api.ApRules{
			Identity: &api.Ref{
				ID:   identity.ID,
				Name: identity.Name,
			},
			Labels: api.InExList{
				Included: []string{"konveyor.io/include=java"},
				Excluded: []string{"konveyor.io/exclude=test"},
			},
			Repository: &api.Repository{
				URL:  "https://github.com/konveyor/rulesets.git",
				Path: "default/generated/camel3",
			},
			Targets: []api.ApTargetRef{
				{ID: 2, Name: "Containerization"},
				{ID: 6, Name: "OpenJDK", Selection: "konveyor.io/target=openjdk17"},
			},
		},
	}

	// CREATE: Create the profile
	err = client.AnalysisProfile.Create(profile)
	g.Expect(err).To(BeNil())
	g.Expect(profile.ID).NotTo(BeZero())

	defer func() {
		_ = client.AnalysisProfile.Delete(profile.ID)
	}()

	// GET: Retrieve the profile and verify it matches
	retrieved, err := client.AnalysisProfile.Get(profile.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	g.Expect(retrieved.ID).To(Equal(profile.ID))
	g.Expect(retrieved.Name).To(Equal(profile.Name))
	g.Expect(retrieved.Description).To(Equal(profile.Description))
	g.Expect(retrieved.Mode.WithDeps).To(Equal(profile.Mode.WithDeps))
	g.Expect(retrieved.Scope.WithKnownLibs).To(Equal(profile.Scope.WithKnownLibs))
	g.Expect(retrieved.Scope.Packages.Included).To(Equal(profile.Scope.Packages.Included))
	g.Expect(retrieved.Scope.Packages.Excluded).To(Equal(profile.Scope.Packages.Excluded))
	g.Expect(retrieved.Rules.Labels.Included).To(Equal(profile.Rules.Labels.Included))
	g.Expect(retrieved.Rules.Labels.Excluded).To(Equal(profile.Rules.Labels.Excluded))
	g.Expect(retrieved.Rules.Repository).NotTo(BeNil())
	g.Expect(retrieved.Rules.Repository.URL).To(Equal(profile.Rules.Repository.URL))
	g.Expect(retrieved.Rules.Repository.Path).To(Equal(profile.Rules.Repository.Path))
	g.Expect(retrieved.Rules.Targets).To(HaveLen(len(profile.Rules.Targets)))
	for i, target := range retrieved.Rules.Targets {
		g.Expect(target.ID).To(Equal(profile.Rules.Targets[i].ID))
		g.Expect(target.Name).To(Equal(profile.Rules.Targets[i].Name))
	}
	g.Expect(retrieved.Rules.Identity).NotTo(BeNil())
	g.Expect(retrieved.Rules.Identity.ID).To(Equal(identity.ID))

	// UPDATE: Modify the profile
	profile.Name = "Updated Test Profile"
	profile.Description = "This is an updated test analysis profile"
	profile.Mode.WithDeps = false
	profile.Scope.WithKnownLibs = false
	profile.Scope.Packages.Included = []string{"com.example.updated1"}
	profile.Scope.Packages.Excluded = []string{"com.example.updated2", "com.example.updated3"}
	profile.Rules.Labels.Included = []string{"konveyor.io/updated=true"}
	profile.Rules.Labels.Excluded = []string{}

	err = client.AnalysisProfile.Update(profile)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.AnalysisProfile.Get(profile.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.ID).To(Equal(profile.ID))
	g.Expect(updated.Name).To(Equal(profile.Name))
	g.Expect(updated.Description).To(Equal(profile.Description))
	g.Expect(updated.Mode.WithDeps).To(Equal(profile.Mode.WithDeps))
	g.Expect(updated.Scope.WithKnownLibs).To(Equal(profile.Scope.WithKnownLibs))
	g.Expect(updated.Scope.Packages.Included).To(Equal(profile.Scope.Packages.Included))
	g.Expect(updated.Scope.Packages.Excluded).To(Equal(profile.Scope.Packages.Excluded))
	g.Expect(updated.Rules.Labels.Included).To(Equal(profile.Rules.Labels.Included))
	g.Expect(updated.Rules.Labels.Excluded).To(Equal(profile.Rules.Labels.Excluded))

	// DELETE: Remove the profile
	err = client.AnalysisProfile.Delete(profile.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.AnalysisProfile.Get(profile.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
