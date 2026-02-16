package resource

import (
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/onsi/gomega"
)

// TestAnalysisProfile_With_Comprehensive tests the AnalysisProfile.With() method with all fields
func TestAnalysisProfile_With_Comprehensive(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.AnalysisProfile{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
			UpdateUser: "user2",
			CreateTime: time.Now(),
		},
		Name:          "test-profile",
		Description:   "test description",
		WithDeps:      true,
		WithKnownLibs: true,
		Packages: model.InExList{
			Included: []string{"pkg1", "pkg2"},
			Excluded: []string{"pkg3"},
		},
		Labels: model.InExList{
			Included: []string{"label1"},
			Excluded: []string{"label2", "label3"},
		},
		Repository: model.Repository{
			Kind:   "git",
			URL:    "https://github.com/test/repo",
			Branch: "main",
			Tag:    "v1.0.0",
			Path:   "/path/to/rules",
		},
		Targets: []model.Target{
			{
				Model: model.Model{ID: 10},
				Name:  "target1",
			},
			{
				Model: model.Model{ID: 20},
				Name:  "target2",
			},
		},
		Files: []json.Ref{
			{ID: 100, Name: "file1"},
			{ID: 200, Name: "file2"},
		},
	}

	r := &AnalysisProfile{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.CreateUser).To(gomega.Equal("user1"))
	g.Expect(r.UpdateUser).To(gomega.Equal("user2"))
	g.Expect(r.Name).To(gomega.Equal("test-profile"))
	g.Expect(r.Description).To(gomega.Equal("test description"))
	g.Expect(r.Mode.WithDeps).To(gomega.Equal(true))
	g.Expect(r.Scope.WithKnownLibs).To(gomega.Equal(true))
	g.Expect(r.Scope.Packages.Included).To(gomega.Equal([]string{"pkg1", "pkg2"}))
	g.Expect(r.Scope.Packages.Excluded).To(gomega.Equal([]string{"pkg3"}))
	g.Expect(r.Rules.Labels.Included).To(gomega.Equal([]string{"label1"}))
	g.Expect(r.Rules.Labels.Excluded).To(gomega.Equal([]string{"label2", "label3"}))
	g.Expect(r.Rules.Repository).ToNot(gomega.BeNil())
	g.Expect(r.Rules.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(r.Rules.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(r.Rules.Repository.Branch).To(gomega.Equal("main"))
	g.Expect(r.Rules.Repository.Tag).To(gomega.Equal("v1.0.0"))
	g.Expect(r.Rules.Repository.Path).To(gomega.Equal("/path/to/rules"))
	g.Expect(len(r.Rules.Targets)).To(gomega.Equal(2))
	g.Expect(r.Rules.Targets[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(r.Rules.Targets[0].Name).To(gomega.Equal("target1"))
	g.Expect(r.Rules.Targets[1].ID).To(gomega.Equal(uint(20)))
	g.Expect(r.Rules.Targets[1].Name).To(gomega.Equal("target2"))
	g.Expect(len(r.Rules.Files)).To(gomega.Equal(2))
	g.Expect(r.Rules.Files[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Rules.Files[0].Name).To(gomega.Equal("file1"))
	g.Expect(r.Rules.Files[1].ID).To(gomega.Equal(uint(200)))
	g.Expect(r.Rules.Files[1].Name).To(gomega.Equal("file2"))
}

// TestAnalysisProfile_With_NilRepository tests the AnalysisProfile.With() method with nil repository
func TestAnalysisProfile_With_NilRepository(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.AnalysisProfile{
		Model: model.Model{
			ID: 1,
		},
		Name:       "test-profile",
		Repository: model.Repository{},
	}

	r := &AnalysisProfile{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-profile"))
	g.Expect(r.Rules.Repository).To(gomega.BeNil())
}

// TestAnalysisProfile_With_EmptySlices tests the AnalysisProfile.With() method with empty slices
func TestAnalysisProfile_With_EmptySlices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.AnalysisProfile{
		Model: model.Model{
			ID: 1,
		},
		Name:    "test-profile",
		Targets: []model.Target{},
		Files:   []json.Ref{},
	}

	r := &AnalysisProfile{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-profile"))
	g.Expect(len(r.Rules.Targets)).To(gomega.Equal(0))
	g.Expect(len(r.Rules.Files)).To(gomega.Equal(0))
}

// TestAnalysisProfile_Model_Comprehensive tests the AnalysisProfile.Model() method with all fields
func TestAnalysisProfile_Model_Comprehensive(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	repository := api.Repository{
		Kind:   "git",
		URL:    "https://github.com/test/repo",
		Branch: "main",
		Tag:    "v1.0.0",
		Path:   "/path/to/rules",
	}

	r := &AnalysisProfile{
		Resource:    Resource{ID: 1},
		Name:        "test-profile",
		Description: "test description",
		Mode: api.ApMode{
			WithDeps: true,
		},
		Scope: api.ApScope{
			WithKnownLibs: true,
			Packages: api.InExList{
				Included: []string{"pkg1", "pkg2"},
				Excluded: []string{"pkg3"},
			},
		},
		Rules: api.ApRules{
			Labels: api.InExList{
				Included: []string{"label1"},
				Excluded: []string{"label2", "label3"},
			},
			Repository: &repository,
			Targets: []api.ApTargetRef{
				{ID: 10, Name: "target1", Selection: "Label4"},
				{ID: 20, Name: "target2"},
			},
			Files: []Ref{
				{ID: 100, Name: "file1"},
				{ID: 200, Name: "file2"},
			},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-profile"))
	g.Expect(m.Description).To(gomega.Equal("test description"))
	g.Expect(m.WithDeps).To(gomega.Equal(true))
	g.Expect(m.WithKnownLibs).To(gomega.Equal(true))
	g.Expect(m.Packages.Included).To(gomega.Equal([]string{"pkg1", "pkg2"}))
	g.Expect(m.Packages.Excluded).To(gomega.Equal([]string{"pkg3"}))
	g.Expect(m.Labels.Included).To(gomega.Equal([]string{"label1"}))
	g.Expect(m.Labels.Excluded).To(gomega.Equal([]string{"label2", "label3"}))
	g.Expect(m.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(m.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(m.Repository.Branch).To(gomega.Equal("main"))
	g.Expect(m.Repository.Tag).To(gomega.Equal("v1.0.0"))
	g.Expect(m.Repository.Path).To(gomega.Equal("/path/to/rules"))
	g.Expect(len(m.Targets)).To(gomega.Equal(2))
	g.Expect(m.Targets[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(m.Targets[1].ID).To(gomega.Equal(uint(20)))
	g.Expect(len(m.Files)).To(gomega.Equal(2))
	g.Expect(m.Files[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(m.Files[0].Name).To(gomega.Equal("file1"))
	g.Expect(m.Files[1].ID).To(gomega.Equal(uint(200)))
	g.Expect(m.Files[1].Name).To(gomega.Equal("file2"))
}

// TestAnalysisProfile_Model_NilRepository tests the AnalysisProfile.Model() method with nil repository
func TestAnalysisProfile_Model_NilRepository(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &AnalysisProfile{
		Resource: Resource{ID: 1},
		Name:     "test-profile",
		Rules: api.ApRules{
			Repository: nil,
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-profile"))
	g.Expect(m.Repository).To(gomega.Equal(model.Repository{}))
}

// TestAnalysisProfile_Model_EmptySlices tests the AnalysisProfile.Model() method with empty slices
func TestAnalysisProfile_Model_EmptySlices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &AnalysisProfile{
		Resource: Resource{ID: 1},
		Name:     "test-profile",
		Rules: api.ApRules{
			Targets: []api.ApTargetRef{},
			Files:   []Ref{},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-profile"))
	g.Expect(len(m.Targets)).To(gomega.Equal(0))
	g.Expect(len(m.Files)).To(gomega.Equal(0))
}

// TestAnalysisProfile_RoundTrip tests the round-trip conversion of AnalysisProfile
func TestAnalysisProfile_RoundTrip(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	original := &model.AnalysisProfile{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
			UpdateUser: "user2",
			CreateTime: time.Now(),
		},
		Name:          "test-profile",
		Description:   "test description",
		WithDeps:      true,
		WithKnownLibs: true,
		Packages: model.InExList{
			Included: []string{"pkg1", "pkg2"},
			Excluded: []string{"pkg3"},
		},
		Labels: model.InExList{
			Included: []string{"label1"},
			Excluded: []string{"label2", "label3"},
		},
		Repository: model.Repository{
			Kind:   "git",
			URL:    "https://github.com/test/repo",
			Branch: "main",
			Tag:    "v1.0.0",
			Path:   "/path/to/rules",
		},
		Targets: []model.Target{
			{
				Model: model.Model{ID: 10},
				Name:  "target1",
			},
			{
				Model: model.Model{ID: 20},
				Name:  "target2",
			},
		},
		Files: []json.Ref{
			{ID: 100, Name: "file1"},
			{ID: 200, Name: "file2"},
		},
	}

	// Convert model to resource
	r := &AnalysisProfile{}
	r.With(original)

	// Convert resource back to model
	result := r.Model()

	// Verify all fields match
	g.Expect(result.ID).To(gomega.Equal(original.ID))
	g.Expect(result.Name).To(gomega.Equal(original.Name))
	g.Expect(result.Description).To(gomega.Equal(original.Description))
	g.Expect(result.WithDeps).To(gomega.Equal(original.WithDeps))
	g.Expect(result.WithKnownLibs).To(gomega.Equal(original.WithKnownLibs))
	g.Expect(result.Packages.Included).To(gomega.Equal(original.Packages.Included))
	g.Expect(result.Packages.Excluded).To(gomega.Equal(original.Packages.Excluded))
	g.Expect(result.Labels.Included).To(gomega.Equal(original.Labels.Included))
	g.Expect(result.Labels.Excluded).To(gomega.Equal(original.Labels.Excluded))
	g.Expect(result.Repository.Kind).To(gomega.Equal(original.Repository.Kind))
	g.Expect(result.Repository.URL).To(gomega.Equal(original.Repository.URL))
	g.Expect(result.Repository.Branch).To(gomega.Equal(original.Repository.Branch))
	g.Expect(result.Repository.Tag).To(gomega.Equal(original.Repository.Tag))
	g.Expect(result.Repository.Path).To(gomega.Equal(original.Repository.Path))
	g.Expect(len(result.Targets)).To(gomega.Equal(len(original.Targets)))
	for i := range result.Targets {
		g.Expect(result.Targets[i].ID).To(gomega.Equal(original.Targets[i].ID))
	}
	g.Expect(len(result.Files)).To(gomega.Equal(len(original.Files)))
	for i := range result.Files {
		g.Expect(result.Files[i].ID).To(gomega.Equal(original.Files[i].ID))
		g.Expect(result.Files[i].Name).To(gomega.Equal(original.Files[i].Name))
	}
}

// TestAnalysisProfile_RoundTrip_NilValues tests the round-trip conversion with nil values
func TestAnalysisProfile_RoundTrip_NilValues(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	original := &model.AnalysisProfile{
		Model: model.Model{
			ID: 1,
		},
		Name:       "test-profile",
		Repository: model.Repository{},
		Targets:    []model.Target{},
		Files:      []json.Ref{},
	}

	// Convert model to resource
	r := &AnalysisProfile{}
	r.With(original)

	// Convert resource back to model
	result := r.Model()

	// Verify fields match
	g.Expect(result.ID).To(gomega.Equal(original.ID))
	g.Expect(result.Name).To(gomega.Equal(original.Name))
	g.Expect(result.Repository).To(gomega.Equal(model.Repository{}))
	g.Expect(len(result.Targets)).To(gomega.Equal(0))
	g.Expect(len(result.Files)).To(gomega.Equal(0))
}

// TestAnalysisProfile_RoundTrip_EmptyInExLists tests round-trip with empty InExLists
func TestAnalysisProfile_RoundTrip_EmptyInExLists(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	original := &model.AnalysisProfile{
		Model: model.Model{
			ID: 1,
		},
		Name: "test-profile",
		Packages: model.InExList{
			Included: nil,
			Excluded: nil,
		},
		Labels: model.InExList{
			Included: nil,
			Excluded: nil,
		},
	}

	// Convert model to resource
	r := &AnalysisProfile{}
	r.With(original)

	// Convert resource back to model
	result := r.Model()

	// Verify fields match
	g.Expect(result.ID).To(gomega.Equal(original.ID))
	g.Expect(result.Name).To(gomega.Equal(original.Name))
	g.Expect(result.Packages.Included).To(gomega.BeNil())
	g.Expect(result.Packages.Excluded).To(gomega.BeNil())
	g.Expect(result.Labels.Included).To(gomega.BeNil())
	g.Expect(result.Labels.Excluded).To(gomega.BeNil())
}
