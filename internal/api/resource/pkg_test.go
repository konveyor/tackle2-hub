package resource

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/api/jsd"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestTag_With tests the Tag.With() method
func TestTag_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Tag{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
			UpdateUser: "user2",
			CreateTime: time.Now(),
		},
		Name:       "test-tag",
		CategoryID: 10,
		Category: model.TagCategory{
			Model: model.Model{ID: 10},
			Name:  "test-category",
		},
	}

	r := &Tag{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.CreateUser).To(gomega.Equal("user1"))
	g.Expect(r.UpdateUser).To(gomega.Equal("user2"))
	g.Expect(r.Name).To(gomega.Equal("test-tag"))
	g.Expect(r.Category.ID).To(gomega.Equal(uint(10)))
	g.Expect(r.Category.Name).To(gomega.Equal("test-category"))
}

// TestTag_Model tests the Tag.Model() method
func TestTag_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Tag{
		Resource: Resource{ID: 1},
		Name:     "test-tag",
		Category: Ref{ID: 10, Name: "test-category"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-tag"))
	g.Expect(m.CategoryID).To(gomega.Equal(uint(10)))
}

// TestBusinessService_With tests the BusinessService.With() method
func TestBusinessService_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	stakeholderID := uint(5)
	m := &model.BusinessService{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:          "test-service",
		Description:   "test description",
		StakeholderID: &stakeholderID,
		Stakeholder: &model.Stakeholder{
			Model: model.Model{ID: 5},
			Name:  "test-stakeholder",
		},
	}

	r := &BusinessService{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-service"))
	g.Expect(r.Description).To(gomega.Equal("test description"))
	g.Expect(r.Stakeholder).ToNot(gomega.BeNil())
	g.Expect(r.Stakeholder.ID).To(gomega.Equal(uint(5)))
	g.Expect(r.Stakeholder.Name).To(gomega.Equal("test-stakeholder"))
}

// TestBusinessService_Model tests the BusinessService.Model() method
func TestBusinessService_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &BusinessService{
		Resource:    Resource{ID: 1},
		Name:        "test-service",
		Description: "test description",
		Stakeholder: &Ref{ID: 5, Name: "test-stakeholder"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-service"))
	g.Expect(m.Description).To(gomega.Equal("test description"))
	g.Expect(m.StakeholderID).ToNot(gomega.BeNil())
	g.Expect(*m.StakeholderID).To(gomega.Equal(uint(5)))
}

// TestBusinessService_Model_NilStakeholder tests BusinessService.Model() with nil stakeholder
func TestBusinessService_Model_NilStakeholder(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &BusinessService{
		Resource:    Resource{ID: 1},
		Name:        "test-service",
		Description: "test description",
		Stakeholder: nil,
	}

	m := r.Model()

	g.Expect(m.StakeholderID).To(gomega.BeNil())
}

// TestJobFunction_With tests the JobFunction.With() method
func TestJobFunction_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.JobFunction{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name: "Developer",
		Stakeholders: []model.Stakeholder{
			{Model: model.Model{ID: 1}, Name: "Alice"},
			{Model: model.Model{ID: 2}, Name: "Bob"},
		},
	}

	r := &JobFunction{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Developer"))
	g.Expect(len(r.Stakeholders)).To(gomega.Equal(2))
	g.Expect(r.Stakeholders[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Stakeholders[0].Name).To(gomega.Equal("Alice"))
	g.Expect(r.Stakeholders[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(r.Stakeholders[1].Name).To(gomega.Equal("Bob"))
}

// TestJobFunction_Model tests the JobFunction.Model() method
func TestJobFunction_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &JobFunction{
		Resource: Resource{ID: 1},
		Name:     "Developer",
		Stakeholders: []Ref{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Developer"))
}

// TestStakeholder_With tests the Stakeholder.With() method
func TestStakeholder_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	jobFuncID := uint(3)
	m := &model.Stakeholder{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:          "John Doe",
		Email:         "john@example.com",
		JobFunctionID: &jobFuncID,
		JobFunction: &model.JobFunction{
			Model: model.Model{ID: 3},
			Name:  "Manager",
		},
		Groups: []model.StakeholderGroup{
			{Model: model.Model{ID: 1}, Name: "Group1"},
			{Model: model.Model{ID: 2}, Name: "Group2"},
		},
		BusinessServices: []model.BusinessService{
			{Model: model.Model{ID: 10}, Name: "Service1"},
		},
		Owns: []model.Application{
			{Model: model.Model{ID: 100}, Name: "App1"},
		},
		Contributes: []model.Application{
			{Model: model.Model{ID: 200}, Name: "App2"},
		},
		MigrationWaves: []model.MigrationWave{
			{Model: model.Model{ID: 50}, Name: "Wave1"},
		},
	}

	r := &Stakeholder{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("John Doe"))
	g.Expect(r.Email).To(gomega.Equal("john@example.com"))
	g.Expect(r.JobFunction).ToNot(gomega.BeNil())
	g.Expect(r.JobFunction.ID).To(gomega.Equal(uint(3)))
	g.Expect(r.JobFunction.Name).To(gomega.Equal("Manager"))
	g.Expect(len(r.Groups)).To(gomega.Equal(2))
	g.Expect(r.Groups[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Groups[0].Name).To(gomega.Equal("Group1"))
	g.Expect(len(r.BusinessServices)).To(gomega.Equal(1))
	g.Expect(r.BusinessServices[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(len(r.Owns)).To(gomega.Equal(1))
	g.Expect(r.Owns[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(len(r.Contributes)).To(gomega.Equal(1))
	g.Expect(r.Contributes[0].ID).To(gomega.Equal(uint(200)))
	g.Expect(len(r.MigrationWaves)).To(gomega.Equal(1))
	g.Expect(r.MigrationWaves[0].ID).To(gomega.Equal(uint(50)))
}

// TestStakeholder_Model tests the Stakeholder.Model() method
func TestStakeholder_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Stakeholder{
		Resource:    Resource{ID: 1},
		Name:        "John Doe",
		Email:       "john@example.com",
		JobFunction: &Ref{ID: 3, Name: "Manager"},
		Groups: []Ref{
			{ID: 1, Name: "Group1"},
			{ID: 2, Name: "Group2"},
		},
		BusinessServices: []Ref{
			{ID: 10, Name: "Service1"},
		},
		Owns: []Ref{
			{ID: 100, Name: "App1"},
		},
		Contributes: []Ref{
			{ID: 200, Name: "App2"},
		},
		MigrationWaves: []Ref{
			{ID: 50, Name: "Wave1"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("John Doe"))
	g.Expect(m.Email).To(gomega.Equal("john@example.com"))
	g.Expect(m.JobFunctionID).ToNot(gomega.BeNil())
	g.Expect(*m.JobFunctionID).To(gomega.Equal(uint(3)))
	g.Expect(len(m.Groups)).To(gomega.Equal(2))
	g.Expect(m.Groups[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Groups[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(len(m.BusinessServices)).To(gomega.Equal(1))
	g.Expect(m.BusinessServices[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(len(m.Owns)).To(gomega.Equal(1))
	g.Expect(m.Owns[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(len(m.Contributes)).To(gomega.Equal(1))
	g.Expect(m.Contributes[0].ID).To(gomega.Equal(uint(200)))
	g.Expect(len(m.MigrationWaves)).To(gomega.Equal(1))
	g.Expect(m.MigrationWaves[0].ID).To(gomega.Equal(uint(50)))
}

// TestStakeholder_Model_NilJobFunction tests Stakeholder.Model() with nil job function
func TestStakeholder_Model_NilJobFunction(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Stakeholder{
		Resource:    Resource{ID: 1},
		Name:        "John Doe",
		Email:       "john@example.com",
		JobFunction: nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("John Doe"))
	g.Expect(m.Email).To(gomega.Equal("john@example.com"))
	g.Expect(m.JobFunctionID).To(gomega.BeNil())
}

// TestTagCategory_With tests the TagCategory.With() method
func TestTagCategory_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.TagCategory{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:  "test-category",
		Color: "#FF0000",
		Tags: []model.Tag{
			{Model: model.Model{ID: 1}, Name: "tag1"},
			{Model: model.Model{ID: 2}, Name: "tag2"},
		},
	}

	r := &TagCategory{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-category"))
	g.Expect(r.Color).To(gomega.Equal("#FF0000"))
	g.Expect(len(r.Tags)).To(gomega.Equal(2))
	g.Expect(r.Tags[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Tags[0].Name).To(gomega.Equal("tag1"))
	g.Expect(r.Tags[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(r.Tags[1].Name).To(gomega.Equal("tag2"))
}

// TestTagCategory_Model tests the TagCategory.Model() method
func TestTagCategory_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &TagCategory{
		Resource: Resource{ID: 1},
		Name:     "test-category",
		Color:    "#FF0000",
		Tags: []Ref{
			{ID: 1, Name: "tag1"},
			{ID: 2, Name: "tag2"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-category"))
	g.Expect(m.Color).To(gomega.Equal("#FF0000"))
}

// TestPlatform_With tests the Platform.With() method
func TestPlatform_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	identityID := uint(7)
	m := &model.Platform{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Kind:       "kubernetes",
		Name:       "prod-cluster",
		URL:        "https://k8s.example.com",
		IdentityID: &identityID,
		Identity: &model.Identity{
			Model: model.Model{ID: 7},
			Name:  "k8s-creds",
		},
		Applications: []model.Application{
			{Model: model.Model{ID: 100}, Name: "app1"},
			{Model: model.Model{ID: 101}, Name: "app2"},
		},
	}

	r := &Platform{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Kind).To(gomega.Equal("kubernetes"))
	g.Expect(r.Name).To(gomega.Equal("prod-cluster"))
	g.Expect(r.URL).To(gomega.Equal("https://k8s.example.com"))
	g.Expect(r.Identity).ToNot(gomega.BeNil())
	g.Expect(r.Identity.ID).To(gomega.Equal(uint(7)))
	g.Expect(r.Identity.Name).To(gomega.Equal("k8s-creds"))
	g.Expect(len(r.Applications)).To(gomega.Equal(2))
	g.Expect(r.Applications[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Applications[1].ID).To(gomega.Equal(uint(101)))
}

// TestPlatform_Model tests the Platform.Model() method
func TestPlatform_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Platform{
		Resource: Resource{ID: 1},
		Kind:     "kubernetes",
		Name:     "prod-cluster",
		URL:      "https://k8s.example.com",
		Identity: &Ref{ID: 7, Name: "k8s-creds"},
		Applications: []Ref{
			{ID: 100, Name: "app1"},
			{ID: 101, Name: "app2"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("kubernetes"))
	g.Expect(m.Name).To(gomega.Equal("prod-cluster"))
	g.Expect(m.URL).To(gomega.Equal("https://k8s.example.com"))
	g.Expect(m.IdentityID).ToNot(gomega.BeNil())
	g.Expect(*m.IdentityID).To(gomega.Equal(uint(7)))
}

// TestPlatform_Model_NilIdentity tests Platform.Model() with nil identity
func TestPlatform_Model_NilIdentity(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Platform{
		Resource: Resource{ID: 1},
		Kind:     "kubernetes",
		Name:     "prod-cluster",
		URL:      "https://k8s.example.com",
		Identity: nil,
	}

	m := r.Model()

	g.Expect(m.IdentityID).To(gomega.BeNil())
}

// TestProxy_With tests the Proxy.With() method
func TestProxy_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	identityID := uint(5)
	m := &model.Proxy{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Enabled:    true,
		Kind:       "http",
		Host:       "proxy.example.com",
		Port:       8080,
		IdentityID: &identityID,
		Identity: &model.Identity{
			Model: model.Model{ID: 5},
			Name:  "proxy-creds",
		},
		Excluded: []string{"*.internal.com", "localhost"},
	}

	r := &Proxy{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Enabled).To(gomega.Equal(true))
	g.Expect(r.Kind).To(gomega.Equal("http"))
	g.Expect(r.Host).To(gomega.Equal("proxy.example.com"))
	g.Expect(r.Port).To(gomega.Equal(8080))
	g.Expect(r.Identity).ToNot(gomega.BeNil())
	g.Expect(r.Identity.ID).To(gomega.Equal(uint(5)))
	g.Expect(len(r.Excluded)).To(gomega.Equal(2))
	g.Expect(r.Excluded[0]).To(gomega.Equal("*.internal.com"))
	g.Expect(r.Excluded[1]).To(gomega.Equal("localhost"))
}

// TestProxy_With_NilExcluded tests Proxy.With() with nil excluded list
func TestProxy_With_NilExcluded(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Proxy{
		Model: model.Model{ID: 1},
		Host:  "proxy.example.com",
		Port:  8080,
	}

	r := &Proxy{}
	r.With(m)

	g.Expect(r.Excluded).ToNot(gomega.BeNil())
	g.Expect(len(r.Excluded)).To(gomega.Equal(0))
}

// TestProxy_Model tests the Proxy.Model() method
func TestProxy_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Proxy{
		Resource: Resource{ID: 1},
		Enabled:  true,
		Kind:     "https",
		Host:     "proxy.example.com",
		Port:     8443,
		Identity: &Ref{ID: 5, Name: "proxy-creds"},
		Excluded: []string{"*.internal.com", "localhost"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Enabled).To(gomega.Equal(true))
	g.Expect(m.Kind).To(gomega.Equal("https"))
	g.Expect(m.Host).To(gomega.Equal("proxy.example.com"))
	g.Expect(m.Port).To(gomega.Equal(8443))
	g.Expect(m.IdentityID).ToNot(gomega.BeNil())
	g.Expect(*m.IdentityID).To(gomega.Equal(uint(5)))
	g.Expect(len(m.Excluded)).To(gomega.Equal(2))
	g.Expect(m.Excluded[0]).To(gomega.Equal("*.internal.com"))
}

// TestProxy_Model_NilIdentity tests Proxy.Model() with nil identity
func TestProxy_Model_NilIdentity(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Proxy{
		Resource: Resource{ID: 1},
		Enabled:  true,
		Kind:     "http",
		Host:     "proxy.example.com",
		Port:     8080,
		Identity: nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Enabled).To(gomega.Equal(true))
	g.Expect(m.IdentityID).To(gomega.BeNil())
}

// TestDependency_With tests the Dependency.With() method
func TestDependency_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Dependency{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		ToID: 100,
		To: &model.Application{
			Model: model.Model{ID: 100},
			Name:  "app-to",
		},
		FromID: 200,
		From: &model.Application{
			Model: model.Model{ID: 200},
			Name:  "app-from",
		},
	}

	r := &Dependency{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.To.ID).To(gomega.Equal(uint(100)))
	g.Expect(r.To.Name).To(gomega.Equal("app-to"))
	g.Expect(r.From.ID).To(gomega.Equal(uint(200)))
	g.Expect(r.From.Name).To(gomega.Equal("app-from"))
}

// TestDependency_Model tests the Dependency.Model() method
func TestDependency_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Dependency{
		Resource: Resource{ID: 1},
		To:       Ref{ID: 100, Name: "app-to"},
		From:     Ref{ID: 200, Name: "app-from"},
	}

	m := r.Model()

	g.Expect(m.ToID).To(gomega.Equal(uint(100)))
	g.Expect(m.FromID).To(gomega.Equal(uint(200)))
}

// TestSetting_With tests the Setting.With() method
func TestSetting_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Setting{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Key:   "ui.theme",
		Value: "dark",
	}

	r := &Setting{}
	r.With(m)

	g.Expect(r.Key).To(gomega.Equal("ui.theme"))
	g.Expect(r.Value).To(gomega.Equal("dark"))
}

// TestSetting_Model tests the Setting.Model() method
func TestSetting_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Setting{
		Key:   "ui.theme",
		Value: "dark",
	}

	m := r.Model()

	g.Expect(m.Key).To(gomega.Equal("ui.theme"))
	g.Expect(m.Value).To(gomega.Equal("dark"))
}

// TestAssessment_With tests the Assessment.With() method
func TestAssessment_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	appID := uint(100)
	m := &model.Assessment{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		ApplicationID: &appID,
		Application: &model.Application{
			Model: model.Model{ID: 100},
			Name:  "test-app",
		},
	}

	r := &Assessment{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Application).ToNot(gomega.BeNil())
	g.Expect(r.Application.ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Application.Name).To(gomega.Equal("test-app"))
}

// TestAssessment_Model tests the Assessment.Model() method
func TestAssessment_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Assessment{
		Resource:    Resource{ID: 1},
		Application: &Ref{ID: 100, Name: "test-app"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.ApplicationID).ToNot(gomega.BeNil())
	g.Expect(*m.ApplicationID).To(gomega.Equal(uint(100)))
}

// TestAssessment_Model_NilPointers tests Assessment.Model() with nil pointer fields
func TestAssessment_Model_NilPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Assessment{
		Resource:    Resource{ID: 1},
		Application: nil,
		Archetype:   nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.ApplicationID).To(gomega.BeNil())
	g.Expect(m.ArchetypeID).To(gomega.BeNil())
}

// TestReview_With tests the Review.With() method
func TestReview_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	appID := uint(100)
	m := &model.Review{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		ApplicationID: &appID,
		Application: &model.Application{
			Model: model.Model{ID: 100},
			Name:  "test-app",
		},
	}

	r := &Review{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Application).ToNot(gomega.BeNil())
	g.Expect(r.Application.ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Application.Name).To(gomega.Equal("test-app"))
}

// TestReview_Model tests the Review.Model() method
func TestReview_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Review{
		Resource:    Resource{ID: 1},
		Application: &Ref{ID: 100, Name: "test-app"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.ApplicationID).ToNot(gomega.BeNil())
	g.Expect(*m.ApplicationID).To(gomega.Equal(uint(100)))
}

// TestReview_Model_NilPointers tests Review.Model() with nil pointer fields
func TestReview_Model_NilPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Review{
		Resource:    Resource{ID: 1},
		Application: nil,
		Archetype:   nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.ApplicationID).To(gomega.BeNil())
	g.Expect(m.ArchetypeID).To(gomega.BeNil())
}

// TestMigrationWave_With tests the MigrationWave.With() method
func TestMigrationWave_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.MigrationWave{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name: "Wave 1",
		Applications: []model.Application{
			{Model: model.Model{ID: 100}, Name: "app1"},
			{Model: model.Model{ID: 101}, Name: "app2"},
		},
		Stakeholders: []model.Stakeholder{
			{Model: model.Model{ID: 1}, Name: "stakeholder1"},
		},
	}

	r := &MigrationWave{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Wave 1"))
	g.Expect(len(r.Applications)).To(gomega.Equal(2))
	g.Expect(r.Applications[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(len(r.Stakeholders)).To(gomega.Equal(1))
	g.Expect(r.Stakeholders[0].ID).To(gomega.Equal(uint(1)))
}

// TestMigrationWave_Model tests the MigrationWave.Model() method
func TestMigrationWave_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &MigrationWave{
		Resource: Resource{ID: 1},
		Name:     "Wave 1",
		Applications: []Ref{
			{ID: 100, Name: "app1"},
			{ID: 101, Name: "app2"},
		},
		Stakeholders: []Ref{
			{ID: 1, Name: "stakeholder1"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Wave 1"))
	g.Expect(len(m.Applications)).To(gomega.Equal(2))
	g.Expect(m.Applications[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(len(m.Stakeholders)).To(gomega.Equal(1))
	g.Expect(m.Stakeholders[0].ID).To(gomega.Equal(uint(1)))
}

// TestTracker_With tests the Tracker.With() method
func TestTracker_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Tracker{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:       "JIRA Tracker",
		Kind:       "jira",
		URL:        "https://jira.example.com",
		IdentityID: 5,
		Identity: &model.Identity{
			Model: model.Model{ID: 5},
			Name:  "jira-creds",
		},
	}

	r := &Tracker{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("JIRA Tracker"))
	g.Expect(r.Kind).To(gomega.Equal("jira"))
	g.Expect(r.URL).To(gomega.Equal("https://jira.example.com"))
	g.Expect(r.Identity.ID).To(gomega.Equal(uint(5)))
	g.Expect(r.Identity.Name).To(gomega.Equal("jira-creds"))
}

// TestTracker_Model tests the Tracker.Model() method
func TestTracker_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Tracker{
		Resource: Resource{ID: 1},
		Name:     "JIRA Tracker",
		Kind:     "jira",
		URL:      "https://jira.example.com",
		Identity: Ref{ID: 5, Name: "jira-creds"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("JIRA Tracker"))
	g.Expect(m.Kind).To(gomega.Equal("jira"))
	g.Expect(m.URL).To(gomega.Equal("https://jira.example.com"))
	g.Expect(m.IdentityID).To(gomega.Equal(uint(5)))
}

// TestIdentity_With tests the Identity.With() method
func TestIdentity_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Identity{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:        "test-identity",
		Kind:        "basic-auth",
		Description: "Test credentials",
		User:        "testuser",
		Password:    "encrypted-password",
	}

	r := &Identity{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-identity"))
	g.Expect(r.Kind).To(gomega.Equal("basic-auth"))
	g.Expect(r.Description).To(gomega.Equal("Test credentials"))
	g.Expect(r.User).To(gomega.Equal("testuser"))
	g.Expect(r.Password).To(gomega.Equal("encrypted-password"))
}

// TestIdentity_Model tests the Identity.Model() method
func TestIdentity_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Identity{
		Resource:    Resource{ID: 1},
		Name:        "test-identity",
		Kind:        "basic-auth",
		Description: "Test credentials",
		User:        "testuser",
		Password:    "encrypted-password",
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-identity"))
	g.Expect(m.Kind).To(gomega.Equal("basic-auth"))
	g.Expect(m.Description).To(gomega.Equal("Test credentials"))
	g.Expect(m.User).To(gomega.Equal("testuser"))
	g.Expect(m.Password).To(gomega.Equal("encrypted-password"))
}

// TestApplication_With tests the Application.With() method
func TestApplication_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	businessServiceID := uint(10)
	ownerID := uint(20)
	migrationWaveID := uint(30)
	platformID := uint(40)

	m := &model.Application{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:              "test-app",
		Description:       "Test application",
		Comments:          "Some comments",
		Binary:            "app.jar",
		BusinessServiceID: &businessServiceID,
		BusinessService: &model.BusinessService{
			Model: model.Model{ID: 10},
			Name:  "test-service",
		},
		OwnerID: &ownerID,
		Owner: &model.Stakeholder{
			Model: model.Model{ID: 20},
			Name:  "owner-name",
		},
		Contributors: []model.Stakeholder{
			{Model: model.Model{ID: 21}, Name: "contributor1"},
			{Model: model.Model{ID: 22}, Name: "contributor2"},
		},
		MigrationWaveID: &migrationWaveID,
		MigrationWave: &model.MigrationWave{
			Model: model.Model{ID: 30},
			Name:  "wave1",
		},
		PlatformID: &platformID,
		Platform: &model.Platform{
			Model: model.Model{ID: 40},
			Name:  "platform1",
		},
		Assessments: []model.Assessment{
			{Model: model.Model{ID: 100}},
			{Model: model.Model{ID: 101}},
		},
	}

	tags := []AppTag{
		{
			TagID:  1,
			Source: "manual",
			Tag: &model.Tag{
				Model: model.Model{ID: 1},
				Name:  "tag1",
			},
		},
		{
			TagID:  2,
			Source: "auto",
			Tag: &model.Tag{
				Model: model.Model{ID: 2},
				Name:  "tag2",
			},
		},
	}

	identities := []IdentityRef{
		{ID: 1, Role: "source", Name: "identity1"},
		{ID: 2, Role: "maven", Name: "identity2"},
	}

	r := &Application{}
	r.With(m, tags, identities)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("test-app"))
	g.Expect(r.Description).To(gomega.Equal("Test application"))
	g.Expect(r.Comments).To(gomega.Equal("Some comments"))
	g.Expect(r.Binary).To(gomega.Equal("app.jar"))
	g.Expect(r.BusinessService).ToNot(gomega.BeNil())
	g.Expect(r.BusinessService.ID).To(gomega.Equal(uint(10)))
	g.Expect(r.BusinessService.Name).To(gomega.Equal("test-service"))
	g.Expect(r.Owner).ToNot(gomega.BeNil())
	g.Expect(r.Owner.ID).To(gomega.Equal(uint(20)))
	g.Expect(r.Owner.Name).To(gomega.Equal("owner-name"))
	g.Expect(len(r.Contributors)).To(gomega.Equal(2))
	g.Expect(r.Contributors[0].ID).To(gomega.Equal(uint(21)))
	g.Expect(r.Contributors[0].Name).To(gomega.Equal("contributor1"))
	g.Expect(r.MigrationWave).ToNot(gomega.BeNil())
	g.Expect(r.MigrationWave.ID).To(gomega.Equal(uint(30)))
	g.Expect(r.Platform).ToNot(gomega.BeNil())
	g.Expect(r.Platform.ID).To(gomega.Equal(uint(40)))
	g.Expect(len(r.Tags)).To(gomega.Equal(2))
	g.Expect(r.Tags[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Tags[0].Name).To(gomega.Equal("tag1"))
	g.Expect(r.Tags[0].Source).To(gomega.Equal("manual"))
	g.Expect(r.Tags[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(r.Tags[1].Name).To(gomega.Equal("tag2"))
	g.Expect(r.Tags[1].Source).To(gomega.Equal("auto"))
	g.Expect(len(r.Identities)).To(gomega.Equal(2))
	g.Expect(r.Identities[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Identities[0].Role).To(gomega.Equal("source"))
	g.Expect(r.Identities[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(r.Identities[1].Role).To(gomega.Equal("maven"))
	g.Expect(len(r.Assessments)).To(gomega.Equal(2))
	g.Expect(r.Assessments[0].ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Assessments[1].ID).To(gomega.Equal(uint(101)))
}

// TestApplication_With_Repository tests Application.With() with repository
func TestApplication_With_Repository(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Application{
		Model: model.Model{ID: 1},
		Name:  "test-app",
		Repository: model.Repository{
			Kind:   "git",
			URL:    "https://github.com/test/repo",
			Branch: "main",
		},
		Assets: model.Repository{
			Kind: "git",
			URL:  "https://github.com/test/assets",
			Path: "/assets",
		},
	}

	r := &Application{}
	r.With(m, []AppTag{}, []IdentityRef{})

	g.Expect(r.Repository).ToNot(gomega.BeNil())
	g.Expect(r.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(r.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(r.Repository.Branch).To(gomega.Equal("main"))
	g.Expect(r.Assets).ToNot(gomega.BeNil())
	g.Expect(r.Assets.Kind).To(gomega.Equal("git"))
	g.Expect(r.Assets.URL).To(gomega.Equal("https://github.com/test/assets"))
	g.Expect(r.Assets.Path).To(gomega.Equal("/assets"))
}

// TestApplication_Model tests the Application.Model() method
func TestApplication_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Application{
		Resource:    Resource{ID: 1},
		Name:        "test-app",
		Description: "Test application",
		Comments:    "Some comments",
		Binary:      "app.jar",
		BusinessService: &Ref{
			ID:   10,
			Name: "test-service",
		},
		Owner: &Ref{
			ID:   20,
			Name: "owner-name",
		},
		Contributors: []Ref{
			{ID: 21, Name: "contributor1"},
			{ID: 22, Name: "contributor2"},
		},
		MigrationWave: &Ref{
			ID:   30,
			Name: "wave1",
		},
		Platform: &Ref{
			ID:   40,
			Name: "platform1",
		},
		Tags: []TagRef{
			{ID: 1, Name: "tag1"},
			{ID: 2, Name: "tag2"},
		},
		Identities: []IdentityRef{
			{ID: 1, Role: "source"},
			{ID: 2, Role: "maven"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-app"))
	g.Expect(m.Description).To(gomega.Equal("Test application"))
	g.Expect(m.Comments).To(gomega.Equal("Some comments"))
	g.Expect(m.Binary).To(gomega.Equal("app.jar"))
	g.Expect(m.BusinessServiceID).ToNot(gomega.BeNil())
	g.Expect(*m.BusinessServiceID).To(gomega.Equal(uint(10)))
	g.Expect(m.OwnerID).ToNot(gomega.BeNil())
	g.Expect(*m.OwnerID).To(gomega.Equal(uint(20)))
	g.Expect(len(m.Contributors)).To(gomega.Equal(2))
	g.Expect(m.Contributors[0].ID).To(gomega.Equal(uint(21)))
	g.Expect(m.Contributors[1].ID).To(gomega.Equal(uint(22)))
	g.Expect(m.MigrationWaveID).ToNot(gomega.BeNil())
	g.Expect(*m.MigrationWaveID).To(gomega.Equal(uint(30)))
	g.Expect(m.PlatformID).ToNot(gomega.BeNil())
	g.Expect(*m.PlatformID).To(gomega.Equal(uint(40)))
	g.Expect(len(m.Tags)).To(gomega.Equal(2))
	g.Expect(m.Tags[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Tags[1].ID).To(gomega.Equal(uint(2)))
	g.Expect(len(m.Identities)).To(gomega.Equal(2))
	g.Expect(m.Identities[0].ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Identities[1].ID).To(gomega.Equal(uint(2)))
}

// TestApplication_Model_WithRepository tests Application.Model() with repository
func TestApplication_Model_WithRepository(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Application{
		Resource: Resource{ID: 1},
		Name:     "test-app",
		Repository: &Repository{
			Kind:   "git",
			URL:    "https://github.com/test/repo",
			Branch: "main",
			Tag:    "v1.0.0",
		},
		Assets: &Repository{
			Kind: "git",
			URL:  "https://github.com/test/assets",
			Path: "/assets",
		},
	}

	m := r.Model()

	g.Expect(m.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(m.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(m.Repository.Branch).To(gomega.Equal("main"))
	g.Expect(m.Repository.Tag).To(gomega.Equal("v1.0.0"))
	g.Expect(m.Assets.Kind).To(gomega.Equal("git"))
	g.Expect(m.Assets.URL).To(gomega.Equal("https://github.com/test/assets"))
	g.Expect(m.Assets.Path).To(gomega.Equal("/assets"))
}

// TestApplication_Model_NilPointers tests Application.Model() with nil pointer fields
func TestApplication_Model_NilPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Application{
		Resource:        Resource{ID: 1},
		Name:            "test-app",
		Repository:      nil,
		Assets:          nil,
		BusinessService: nil,
		Owner:           nil,
		MigrationWave:   nil,
		Platform:        nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("test-app"))
	g.Expect(m.BusinessServiceID).To(gomega.BeNil())
	g.Expect(m.OwnerID).To(gomega.BeNil())
	g.Expect(m.MigrationWaveID).To(gomega.BeNil())
	g.Expect(m.PlatformID).To(gomega.BeNil())
}

// TestFact_With tests the Fact.With() method
func TestFact_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Fact{
		Key:    "language",
		Source: "analyzer",
		Value:  "Java",
	}

	r := &Fact{}
	r.With(m)

	g.Expect(r.Key).To(gomega.Equal("language"))
	g.Expect(r.Source).To(gomega.Equal("analyzer"))
	g.Expect(r.Value).To(gomega.Equal("Java"))
}

// TestFact_Model tests the Fact.Model() method
func TestFact_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Fact{
		Key:    "language",
		Source: "analyzer",
		Value:  "Java",
	}

	m := r.Model()

	g.Expect(m.Key).To(gomega.Equal("language"))
	g.Expect(m.Source).To(gomega.Equal("analyzer"))
	g.Expect(m.Value).To(gomega.Equal("Java"))
}

// TestGenerator_With tests the Generator.With() method
func TestGenerator_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	identityID := uint(5)
	m := &model.Generator{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Kind:        "ansible",
		Name:        "test-generator",
		Description: "Test generator",
		IdentityID:  &identityID,
		Identity: &model.Identity{
			Model: model.Model{ID: 5},
			Name:  "gen-creds",
		},
		Repository: model.Repository{
			Kind: "git",
			URL:  "https://github.com/test/repo",
		},
		Params: Map{"key1": "value1"},
		Values: Map{"key2": "value2"},
		Profiles: []model.TargetProfile{
			{Model: model.Model{ID: 10}, Name: "profile1"},
			{Model: model.Model{ID: 11}, Name: "profile2"},
		},
	}

	r := &Generator{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Kind).To(gomega.Equal("ansible"))
	g.Expect(r.Name).To(gomega.Equal("test-generator"))
	g.Expect(r.Description).To(gomega.Equal("Test generator"))
	g.Expect(r.Identity).ToNot(gomega.BeNil())
	g.Expect(r.Identity.ID).To(gomega.Equal(uint(5)))
	g.Expect(r.Repository).ToNot(gomega.BeNil())
	g.Expect(r.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(r.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(r.Params).To(gomega.Equal(Map{"key1": "value1"}))
	g.Expect(r.Values).To(gomega.Equal(Map{"key2": "value2"}))
	g.Expect(len(r.Profiles)).To(gomega.Equal(2))
	g.Expect(r.Profiles[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(r.Profiles[1].ID).To(gomega.Equal(uint(11)))
}

// TestGenerator_Model tests the Generator.Model() method
func TestGenerator_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Generator{
		Resource:    Resource{ID: 1},
		Kind:        "ansible",
		Name:        "test-generator",
		Description: "Test generator",
		Identity:    &Ref{ID: 5, Name: "gen-creds"},
		Repository: &Repository{
			Kind: "git",
			URL:  "https://github.com/test/repo",
		},
		Params: Map{"key1": "value1"},
		Values: Map{"key2": "value2"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("ansible"))
	g.Expect(m.Name).To(gomega.Equal("test-generator"))
	g.Expect(m.Description).To(gomega.Equal("Test generator"))
	g.Expect(m.IdentityID).ToNot(gomega.BeNil())
	g.Expect(*m.IdentityID).To(gomega.Equal(uint(5)))
	g.Expect(m.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(m.Repository.URL).To(gomega.Equal("https://github.com/test/repo"))
	g.Expect(m.Params["key1"]).To(gomega.Equal("value1"))
	g.Expect(m.Values["key2"]).To(gomega.Equal("value2"))
}

// TestGenerator_Model_NilPointers tests Generator.Model() with nil pointer fields
func TestGenerator_Model_NilPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Generator{
		Resource:    Resource{ID: 1},
		Kind:        "ansible",
		Name:        "test-generator",
		Description: "Test generator",
		Identity:    nil,
		Repository:  nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("ansible"))
	g.Expect(m.Name).To(gomega.Equal("test-generator"))
	g.Expect(m.IdentityID).To(gomega.BeNil())
}

// TestTicket_With tests the Ticket.With() method
func TestTicket_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	now := time.Now()
	m := &model.Ticket{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Kind:          "jira",
		Reference:     "ISSUE-123",
		Parent:        "PARENT-1",
		Link:          "https://jira.example.com/ISSUE-123",
		Error:         false,
		Message:       "Ticket created",
		Status:        "Open",
		LastUpdated:   now,
		ApplicationID: 100,
		Application: &model.Application{
			Model: model.Model{ID: 100},
			Name:  "test-app",
		},
		TrackerID: 200,
		Tracker: &model.Tracker{
			Model: model.Model{ID: 200},
			Name:  "test-tracker",
		},
		Fields: Map{"priority": "high"},
	}

	r := &Ticket{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Kind).To(gomega.Equal("jira"))
	g.Expect(r.Reference).To(gomega.Equal("ISSUE-123"))
	g.Expect(r.Parent).To(gomega.Equal("PARENT-1"))
	g.Expect(r.Link).To(gomega.Equal("https://jira.example.com/ISSUE-123"))
	g.Expect(r.Error).To(gomega.Equal(false))
	g.Expect(r.Message).To(gomega.Equal("Ticket created"))
	g.Expect(r.Status).To(gomega.Equal("Open"))
	g.Expect(r.LastUpdated).To(gomega.Equal(now))
	g.Expect(r.Application.ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Tracker.ID).To(gomega.Equal(uint(200)))
	g.Expect(r.Fields).To(gomega.Equal(Map{"priority": "high"}))
}

// TestTicket_Model tests the Ticket.Model() method
func TestTicket_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Ticket{
		Resource:    Resource{ID: 1},
		Kind:        "jira",
		Parent:      "PARENT-1",
		Application: Ref{ID: 100, Name: "test-app"},
		Tracker:     Ref{ID: 200, Name: "test-tracker"},
		Fields:      Map{"priority": "high"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("jira"))
	g.Expect(m.Parent).To(gomega.Equal("PARENT-1"))
	g.Expect(m.ApplicationID).To(gomega.Equal(uint(100)))
	g.Expect(m.TrackerID).To(gomega.Equal(uint(200)))
	g.Expect(m.Fields["priority"]).To(gomega.Equal("high"))
}

// TestQuestionnaire_With tests the Questionnaire.With() method
func TestQuestionnaire_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Questionnaire{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:        "Test Questionnaire",
		Description: "Test description",
		Required:    true,
		Sections: []model.Section{
			{Order: 1, Name: "Section 1"},
		},
		Thresholds: model.Thresholds{
			Red:     50,
			Yellow:  30,
			Unknown: 10,
		},
		RiskMessages: model.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
	}

	r := &Questionnaire{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Questionnaire"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Required).To(gomega.Equal(true))
	g.Expect(len(r.Sections)).To(gomega.Equal(1))
	g.Expect(r.Thresholds.Red).To(gomega.Equal(uint(50)))
	g.Expect(r.RiskMessages.Red).To(gomega.Equal("High risk"))
}

// TestQuestionnaire_Model tests the Questionnaire.Model() method
func TestQuestionnaire_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Questionnaire{
		Resource:    Resource{ID: 1},
		Name:        "Test Questionnaire",
		Description: "Test description",
		Required:    true,
		Sections: []api.Section{
			{Order: 1, Name: "Section 1"},
		},
		Thresholds: api.Thresholds{
			Red:     50,
			Yellow:  30,
			Unknown: 10,
		},
		RiskMessages: api.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Questionnaire"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(m.Required).To(gomega.Equal(true))
	g.Expect(len(m.Sections)).To(gomega.Equal(1))
	g.Expect(m.Thresholds.Red).To(gomega.Equal(uint(50)))
	g.Expect(m.RiskMessages.Red).To(gomega.Equal("High risk"))
}

// TestManifest_With tests the Manifest.With() method
func TestManifest_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Manifest{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Content:       Map{"key1": "value1"},
		Secret:        Map{"password": "secret123"},
		ApplicationID: 100,
	}

	r := &Manifest{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Content).To(gomega.Equal(Map{"key1": "value1"}))
	g.Expect(r.Secret).To(gomega.Equal(Map{"password": "secret123"}))
	g.Expect(r.Application.ID).To(gomega.Equal(uint(100)))
}

// TestManifest_Model tests the Manifest.Model() method
func TestManifest_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Manifest{
		Resource:    Resource{ID: 1},
		Content:     Map{"key1": "value1"},
		Secret:      Map{"password": "secret123"},
		Application: Ref{ID: 100},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Content["key1"]).To(gomega.Equal("value1"))
	g.Expect(m.Secret["password"]).To(gomega.Equal("secret123"))
	g.Expect(m.ApplicationID).To(gomega.Equal(uint(100)))
}

// TestStakeholderGroup_With tests the StakeholderGroup.With() method
func TestStakeholderGroup_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.StakeholderGroup{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:        "Test Group",
		Description: "Test description",
		Stakeholders: []model.Stakeholder{
			{Model: model.Model{ID: 10}, Name: "stakeholder1"},
			{Model: model.Model{ID: 11}, Name: "stakeholder2"},
		},
		MigrationWaves: []model.MigrationWave{
			{Model: model.Model{ID: 20}, Name: "wave1"},
		},
	}

	r := &StakeholderGroup{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Group"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(len(r.Stakeholders)).To(gomega.Equal(2))
	g.Expect(r.Stakeholders[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(r.Stakeholders[0].Name).To(gomega.Equal("stakeholder1"))
	g.Expect(len(r.MigrationWaves)).To(gomega.Equal(1))
	g.Expect(r.MigrationWaves[0].ID).To(gomega.Equal(uint(20)))
}

// TestStakeholderGroup_Model tests the StakeholderGroup.Model() method
func TestStakeholderGroup_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &StakeholderGroup{
		Resource:    Resource{ID: 1},
		Name:        "Test Group",
		Description: "Test description",
		Stakeholders: []Ref{
			{ID: 10, Name: "stakeholder1"},
			{ID: 11, Name: "stakeholder2"},
		},
		MigrationWaves: []Ref{
			{ID: 20, Name: "wave1"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Group"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(len(m.Stakeholders)).To(gomega.Equal(2))
	g.Expect(m.Stakeholders[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(len(m.MigrationWaves)).To(gomega.Equal(1))
	g.Expect(m.MigrationWaves[0].ID).To(gomega.Equal(uint(20)))
}

// TestAnalysis_With tests the Analysis.With() method
func TestAnalysis_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Analysis{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		ApplicationID: 100,
		Application: &model.Application{
			Model: model.Model{ID: 100},
			Name:  "test-app",
		},
		Effort:   50,
		Commit:   "abc123",
		Archived: false,
	}

	r := &Analysis{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Application.ID).To(gomega.Equal(uint(100)))
	g.Expect(r.Effort).To(gomega.Equal(50))
	g.Expect(r.Commit).To(gomega.Equal("abc123"))
	g.Expect(r.Archived).To(gomega.Equal(false))
}

// TestAnalysis_Model tests the Analysis.Model() method
func TestAnalysis_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Analysis{
		Effort: 50,
		Commit: "abc123",
	}

	m := r.Model()

	g.Expect(m.Effort).To(gomega.Equal(50))
	g.Expect(m.Commit).To(gomega.Equal("abc123"))
}

// TestInsight_With tests the Insight.With() method
func TestInsight_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Insight{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		AnalysisID:  100,
		RuleSet:     "ruleset1",
		Rule:        "rule1",
		Name:        "Test Insight",
		Description: "Test description",
		Category:    "potential",
		Incidents: []model.Incident{
			{Model: model.Model{ID: 10}, File: "test.java", Line: 42},
		},
		Links: []model.Link{
			{URL: "https://example.com", Title: "Example"},
		},
		Facts:  Map{"key": "value"},
		Labels: []string{"label1", "label2"},
		Effort: 5,
	}

	r := &Insight{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Analysis).To(gomega.Equal(uint(100)))
	g.Expect(r.RuleSet).To(gomega.Equal("ruleset1"))
	g.Expect(r.Rule).To(gomega.Equal("rule1"))
	g.Expect(r.Name).To(gomega.Equal("Test Insight"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Category).To(gomega.Equal("potential"))
	g.Expect(len(r.Incidents)).To(gomega.Equal(1))
	g.Expect(r.Incidents[0].File).To(gomega.Equal("test.java"))
	g.Expect(len(r.Links)).To(gomega.Equal(1))
	g.Expect(r.Effort).To(gomega.Equal(5))
}

// TestInsight_Model tests the Insight.Model() method
func TestInsight_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Insight{
		RuleSet:     "ruleset1",
		Rule:        "rule1",
		Name:        "Test Insight",
		Description: "Test description",
		Category:    "potential",
		Facts:       Map{"key": "value"},
		Labels:      []string{"label1"},
		Effort:      5,
	}

	m := r.Model()

	g.Expect(m.RuleSet).To(gomega.Equal("ruleset1"))
	g.Expect(m.Rule).To(gomega.Equal("rule1"))
	g.Expect(m.Name).To(gomega.Equal("Test Insight"))
	g.Expect(m.Category).To(gomega.Equal("potential"))
	g.Expect(m.Effort).To(gomega.Equal(5))
}

// TestTechDependency_With tests the TechDependency.With() method
func TestTechDependency_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.TechDependency{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		AnalysisID: 100,
		Provider:   "java",
		Name:       "springframework",
		Version:    "5.3.0",
		Indirect:   false,
		SHA:        "abc123def",
		Labels:     []string{"framework", "web"},
	}

	r := &TechDependency{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Analysis).To(gomega.Equal(uint(100)))
	g.Expect(r.Provider).To(gomega.Equal("java"))
	g.Expect(r.Name).To(gomega.Equal("springframework"))
	g.Expect(r.Version).To(gomega.Equal("5.3.0"))
	g.Expect(r.Indirect).To(gomega.Equal(false))
	g.Expect(r.SHA).To(gomega.Equal("abc123def"))
	g.Expect(len(r.Labels)).To(gomega.Equal(2))
}

// TestTechDependency_Model tests the TechDependency.Model() method
func TestTechDependency_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &TechDependency{
		Provider: "java",
		Name:     "springframework",
		Version:  "5.3.0",
		Indirect: false,
		SHA:      "abc123def",
		Labels:   []string{"web", "framework"},
	}

	m := r.Model()

	g.Expect(m.Provider).To(gomega.Equal("java"))
	g.Expect(m.Name).To(gomega.Equal("springframework"))
	g.Expect(m.Version).To(gomega.Equal("5.3.0"))
	g.Expect(m.Indirect).To(gomega.Equal(false))
	g.Expect(m.SHA).To(gomega.Equal("abc123def"))
	g.Expect(len(m.Labels)).To(gomega.Equal(2))
}

// TestIncident_With tests the Incident.With() method
func TestIncident_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Incident{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		InsightID: 100,
		File:      "test.java",
		Line:      42,
		Message:   "Potential issue",
		CodeSnip:  "public void test() {}",
		Facts:     Map{"severity": "high"},
	}

	r := &Incident{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Insight).To(gomega.Equal(uint(100)))
	g.Expect(r.File).To(gomega.Equal("test.java"))
	g.Expect(r.Line).To(gomega.Equal(42))
	g.Expect(r.Message).To(gomega.Equal("Potential issue"))
	g.Expect(r.CodeSnip).To(gomega.Equal("public void test() {}"))
	g.Expect(r.Facts).To(gomega.Equal(Map{"severity": "high"}))
}

// TestIncident_Model tests the Incident.Model() method
func TestIncident_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Incident{
		File:     "test.java",
		Line:     42,
		Message:  "Potential issue",
		CodeSnip: "public void test() {}",
		Facts:    Map{"severity": "high"},
	}

	m := r.Model()

	g.Expect(m.File).To(gomega.Equal("test.java"))
	g.Expect(m.Line).To(gomega.Equal(42))
	g.Expect(m.Message).To(gomega.Equal("Potential issue"))
	g.Expect(m.CodeSnip).To(gomega.Equal("public void test() {}"))
	g.Expect(m.Facts["severity"]).To(gomega.Equal("high"))
}

// TestAnalysisProfile_With tests the AnalysisProfile.With() method
func TestAnalysisProfile_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	identity := model.Identity{}
	identity.ID = 44
	identity.Name = "Test"

	m := &model.AnalysisProfile{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:          "Test Profile",
		Description:   "Test description",
		WithDeps:      true,
		WithKnownLibs: false,
		Packages:      model.InExList{Included: []string{"com.example"}},
		Labels:        model.InExList{Included: []string{"label1"}},
		Repository: model.Repository{
			Kind: "git",
			URL:  "https://github.com/test/rules",
		},
		IdentityID: &identity.ID,
		Identity:   &identity,
		Targets: []model.Target{
			{Model: model.Model{ID: 1}, Name: "target1"},
			{Model: model.Model{ID: 2}, Name: "target2"},
		},
		Selections: []model.TargetSelection{
			{ID: 1, Label: "xx"},
		},
	}

	r := &AnalysisProfile{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Profile"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Mode.WithDeps).To(gomega.Equal(true))
	g.Expect(r.Scope.WithKnownLibs).To(gomega.Equal(false))
	g.Expect(r.Scope.Packages).To(gomega.Equal(api.InExList{Included: []string{"com.example"}}))
	g.Expect(r.Rules.Labels).To(gomega.Equal(api.InExList{Included: []string{"label1"}}))
	g.Expect(r.Rules.Targets).To(gomega.Equal(
		[]api.ApTargetRef{{ID: 1, Name: "target1", Selection: "xx"}, {ID: 2, Name: "target2"}}))
	g.Expect(r.Rules.Identity).To(gomega.Equal(&api.Ref{ID: identity.ID, Name: identity.Name}))
}

// TestAnalysisProfile_Model tests the AnalysisProfile.Model() method
func TestAnalysisProfile_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &AnalysisProfile{
		Name:        "Test Profile",
		Description: "Test description",
		Mode: api.ApMode{
			WithDeps: true,
		},
		Scope: api.ApScope{
			WithKnownLibs: false,
			Packages:      api.InExList{Included: []string{"com.example"}},
		},
		Rules: api.ApRules{
			Targets: []api.ApTargetRef{
				{ID: 1, Selection: "Label1"},
				{ID: 2},
			},
			Labels:   api.InExList{Included: []string{"label1"}},
			Identity: &Ref{ID: 2},
		},
	}

	m := r.Model()

	uint2 := uint(2)
	g.Expect(m.Name).To(gomega.Equal("Test Profile"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(m.WithDeps).To(gomega.Equal(true))
	g.Expect(m.WithKnownLibs).To(gomega.Equal(false))
	g.Expect(m.Packages).To(gomega.Equal(model.InExList{Included: []string{"com.example"}}))
	g.Expect(m.Labels).To(gomega.Equal(model.InExList{Included: []string{"label1"}}))
	g.Expect(m.IdentityID).To(gomega.Equal(&uint2))
	g.Expect(m.Selections).To(gomega.Equal([]model.TargetSelection{{ID: uint(1), Label: "Label1"}}))
}

// TestArchetype_With tests the Archetype.With() method
func TestArchetype_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Archetype{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:        "Test Archetype",
		Description: "Test description",
		Comments:    "Some comments",
		Tags: []model.Tag{
			{Model: model.Model{ID: 10}, Name: "tag1"},
		},
		CriteriaTags: []model.Tag{
			{Model: model.Model{ID: 20}, Name: "criteria1"},
		},
		Stakeholders: []model.Stakeholder{
			{Model: model.Model{ID: 30}, Name: "stakeholder1"},
		},
		StakeholderGroups: []model.StakeholderGroup{
			{Model: model.Model{ID: 40}, Name: "group1"},
		},
	}

	r := &Archetype{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Archetype"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Comments).To(gomega.Equal("Some comments"))
	g.Expect(len(r.Tags)).To(gomega.Equal(1))
	g.Expect(r.Tags[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(len(r.Criteria)).To(gomega.Equal(1))
	g.Expect(r.Criteria[0].ID).To(gomega.Equal(uint(20)))
	g.Expect(len(r.Stakeholders)).To(gomega.Equal(1))
	g.Expect(len(r.StakeholderGroups)).To(gomega.Equal(1))
}

// TestArchetype_Model tests the Archetype.Model() method
func TestArchetype_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Archetype{
		Resource:    Resource{ID: 1},
		Name:        "Test Archetype",
		Description: "Test description",
		Comments:    "Some comments",
		Tags: []TagRef{
			{ID: 10, Name: "tag1"},
		},
		Criteria: []TagRef{
			{ID: 20, Name: "criteria1"},
		},
		Stakeholders: []Ref{
			{ID: 30, Name: "stakeholder1"},
		},
		StakeholderGroups: []Ref{
			{ID: 40, Name: "group1"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Archetype"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(m.Comments).To(gomega.Equal("Some comments"))
	g.Expect(len(m.Tags)).To(gomega.Equal(1))
	g.Expect(m.Tags[0].ID).To(gomega.Equal(uint(10)))
}

// TestTargetProfile_With tests the TargetProfile.With() method
func TestTargetProfile_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	analysisProfileID := uint(50)
	m := &model.TargetProfile{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name: "Test Target Profile",
		Generators: []model.ProfileGenerator{
			{
				Generator: model.Generator{
					Model: model.Model{ID: 10},
					Name:  "generator1",
				},
			},
		},
		AnalysisProfileID: &analysisProfileID,
		AnalysisProfile: &model.AnalysisProfile{
			Model: model.Model{ID: 50},
			Name:  "analysis-profile",
		},
	}

	r := &TargetProfile{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Target Profile"))
	g.Expect(len(r.Generators)).To(gomega.Equal(1))
	g.Expect(r.Generators[0].ID).To(gomega.Equal(uint(10)))
	g.Expect(r.AnalysisProfile).ToNot(gomega.BeNil())
	g.Expect(r.AnalysisProfile.ID).To(gomega.Equal(uint(50)))
}

// TestTargetProfile_Model tests the TargetProfile.Model() method
func TestTargetProfile_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &TargetProfile{
		Resource: Resource{ID: 1},
		Name:     "Test Target Profile",
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Target Profile"))
}

// TestTargetProfile_Model_NilAnalysisProfile tests TargetProfile.Model() with nil analysis profile
func TestTargetProfile_Model_NilAnalysisProfile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &TargetProfile{
		Resource:        Resource{ID: 1},
		Name:            "Test Target Profile",
		AnalysisProfile: nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Target Profile"))
	g.Expect(m.AnalysisProfileID).To(gomega.BeNil())
}

// TestTarget_With tests the Target.With() method
func TestTarget_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Target{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:        "Test Target",
		Description: "Test description",
		Provider:    "java",
		Choice:      true,
		ImageID:     100,
		Image: &model.File{
			Model: model.Model{ID: 100},
			Name:  "target-image",
		},
		Labels: []model.TargetLabel{
			{Name: "label1", Label: "value1"},
		},
	}

	r := &Target{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Target"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Provider).To(gomega.Equal("java"))
	g.Expect(r.Choice).To(gomega.Equal(true))
	g.Expect(r.Custom).To(gomega.Equal(true))
	g.Expect(r.Image.ID).To(gomega.Equal(uint(100)))
	g.Expect(len(r.Labels)).To(gomega.Equal(1))
}

// TestTarget_Model tests the Target.Model() method
func TestTarget_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Target{
		Resource:    Resource{ID: 1},
		Name:        "Test Target",
		Description: "Test description",
		Provider:    "java",
		Choice:      true,
		Image:       Ref{ID: 100, Name: "target-image"},
		Labels: []api.TargetLabel{
			{Name: "label1", Label: "value1"},
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Target"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(m.Provider).To(gomega.Equal("java"))
	g.Expect(m.Choice).To(gomega.Equal(true))
	g.Expect(m.ImageID).To(gomega.Equal(uint(100)))
}

// TestRuleSet_With tests the RuleSet.With() method
func TestRuleSet_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	identityID := uint(5)
	m := &model.RuleSet{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Kind:        "custom",
		Name:        "Test RuleSet",
		Description: "Test description",
		IdentityID:  &identityID,
		Identity: &model.Identity{
			Model: model.Model{ID: 5},
			Name:  "ruleset-creds",
		},
		Repository: model.Repository{
			Kind: "git",
			URL:  "https://github.com/test/rules",
		},
		Rules: []model.Rule{
			{Model: model.Model{ID: 10}, Name: "rule1"},
		},
		DependsOn: []model.RuleSet{
			{Model: model.Model{ID: 20}, Name: "dependency1"},
		},
	}

	r := &RuleSet{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Kind).To(gomega.Equal("custom"))
	g.Expect(r.Name).To(gomega.Equal("Test RuleSet"))
	g.Expect(r.Description).To(gomega.Equal("Test description"))
	g.Expect(r.Identity).ToNot(gomega.BeNil())
	g.Expect(r.Identity.ID).To(gomega.Equal(uint(5)))
	g.Expect(r.Repository).ToNot(gomega.BeNil())
	g.Expect(r.Repository.Kind).To(gomega.Equal("git"))
	g.Expect(len(r.Rules)).To(gomega.Equal(1))
	g.Expect(len(r.DependsOn)).To(gomega.Equal(1))
}

// TestRuleSet_Model tests the RuleSet.Model() method
func TestRuleSet_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &RuleSet{
		Resource:    Resource{ID: 1},
		Kind:        "custom",
		Name:        "Test RuleSet",
		Description: "Test description",
		Identity:    &Ref{ID: 5, Name: "ruleset-creds"},
		Repository: &Repository{
			Kind: "git",
			URL:  "https://github.com/test/rules",
		},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("custom"))
	g.Expect(m.Name).To(gomega.Equal("Test RuleSet"))
	g.Expect(m.Description).To(gomega.Equal("Test description"))
	g.Expect(m.IdentityID).ToNot(gomega.BeNil())
	g.Expect(*m.IdentityID).To(gomega.Equal(uint(5)))
}

// TestRuleSet_Model_NilPointers tests RuleSet.Model() with nil pointer fields
func TestRuleSet_Model_NilPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &RuleSet{
		Resource:    Resource{ID: 1},
		Kind:        "custom",
		Name:        "Test RuleSet",
		Description: "Test description",
		Identity:    nil,
		Repository:  nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Kind).To(gomega.Equal("custom"))
	g.Expect(m.Name).To(gomega.Equal("Test RuleSet"))
	g.Expect(m.IdentityID).To(gomega.BeNil())
}

// TestRule_With tests the Rule.With() method
func TestRule_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	fileID := uint(100)
	m := &model.Rule{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:   "Test Rule",
		Labels: []string{"label1", "label2"},
		FileID: &fileID,
		File: &model.File{
			Model: model.Model{ID: 100},
			Name:  "rule-file",
		},
	}

	r := &Rule{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Test Rule"))
	g.Expect(len(r.Labels)).To(gomega.Equal(2))
	g.Expect(r.File).ToNot(gomega.BeNil())
	g.Expect(r.File.ID).To(gomega.Equal(uint(100)))
}

// TestRule_Model tests the Rule.Model() method
func TestRule_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Rule{
		Resource: Resource{ID: 1},
		Name:     "Test Rule",
		Labels:   []string{"label1", "label2"},
		File:     &Ref{ID: 100, Name: "rule-file"},
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Rule"))
	g.Expect(len(m.Labels)).To(gomega.Equal(2))
	g.Expect(m.FileID).ToNot(gomega.BeNil())
	g.Expect(*m.FileID).To(gomega.Equal(uint(100)))
}

// TestRule_Model_NilFile tests Rule.Model() with nil file
func TestRule_Model_NilFile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Rule{
		Resource: Resource{ID: 1},
		Name:     "Test Rule",
		Labels:   []string{"label1"},
		File:     nil,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Test Rule"))
	g.Expect(m.FileID).To(gomega.BeNil())
}

// TestConvertContainer tests the convertContainer function
func TestConvertContainer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	pullPolicy := core.PullAlways
	termMsgPolicy := core.TerminationMessageReadFile

	src := core.Container{
		Name:                     "test-container",
		Image:                    "test-image:latest",
		Command:                  []string{"/bin/sh"},
		Args:                     []string{"-c", "echo hello"},
		WorkingDir:               "/app",
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: termMsgPolicy,
		ImagePullPolicy:          pullPolicy,
		Stdin:                    true,
		StdinOnce:                false,
		TTY:                      true,
		Ports: []core.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 8080,
				Protocol:      core.ProtocolTCP,
			},
		},
		Env: []core.EnvVar{
			{
				Name:  "ENV_VAR",
				Value: "value",
			},
		},
		Resources: core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("256Mi"),
			},
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("50m"),
				core.ResourceMemory: resource.MustParse("128Mi"),
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(dst.Name).To(gomega.Equal("test-container"))
	g.Expect(dst.Image).To(gomega.Equal("test-image:latest"))
	g.Expect(dst.Command).To(gomega.Equal([]string{"/bin/sh"}))
	g.Expect(dst.Args).To(gomega.Equal([]string{"-c", "echo hello"}))
	g.Expect(dst.WorkingDir).To(gomega.Equal("/app"))
	g.Expect(dst.TerminationMessagePath).To(gomega.Equal("/dev/termination-log"))
	g.Expect(dst.TerminationMessagePolicy).To(gomega.Equal("File"))
	g.Expect(dst.ImagePullPolicy).To(gomega.Equal("Always"))
	g.Expect(dst.Stdin).To(gomega.Equal(true))
	g.Expect(dst.StdinOnce).To(gomega.Equal(false))
	g.Expect(dst.TTY).To(gomega.Equal(true))
	g.Expect(len(dst.Ports)).To(gomega.Equal(1))
	g.Expect(dst.Ports[0].Name).To(gomega.Equal("http"))
	g.Expect(dst.Ports[0].ContainerPort).To(gomega.Equal(int32(8080)))
	g.Expect(dst.Ports[0].Protocol).To(gomega.Equal("TCP"))
	g.Expect(len(dst.Env)).To(gomega.Equal(1))
	g.Expect(dst.Env[0].Name).To(gomega.Equal("ENV_VAR"))
	g.Expect(dst.Env[0].Value).To(gomega.Equal("value"))
	g.Expect(dst.Resources.Limits["cpu"]).To(gomega.Equal("100m"))
	g.Expect(dst.Resources.Limits["memory"]).To(gomega.Equal("256Mi"))
	g.Expect(dst.Resources.Requests["cpu"]).To(gomega.Equal("50m"))
	g.Expect(dst.Resources.Requests["memory"]).To(gomega.Equal("128Mi"))
}

// TestConvertContainer_WithEnvFrom tests convertContainer with EnvFrom
func TestConvertContainer_WithEnvFrom(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := core.Container{
		Name: "test",
		EnvFrom: []core.EnvFromSource{
			{
				Prefix: "TEST_",
				ConfigMapRef: &core.ConfigMapEnvSource{
					LocalObjectReference: core.LocalObjectReference{Name: "config-map"},
				},
			},
			{
				SecretRef: &core.SecretEnvSource{
					LocalObjectReference: core.LocalObjectReference{Name: "secret"},
				},
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(len(dst.EnvFrom)).To(gomega.Equal(2))
	g.Expect(dst.EnvFrom[0].Prefix).To(gomega.Equal("TEST_"))
	g.Expect(dst.EnvFrom[0].ConfigMapRef).ToNot(gomega.BeNil())
	g.Expect(dst.EnvFrom[0].ConfigMapRef.LocalObjectReference.Name).To(gomega.Equal("config-map"))
	g.Expect(dst.EnvFrom[1].SecretRef).ToNot(gomega.BeNil())
	g.Expect(dst.EnvFrom[1].SecretRef.LocalObjectReference.Name).To(gomega.Equal("secret"))
}

// TestConvertContainer_WithEnvValueFrom tests convertContainer with Env ValueFrom
func TestConvertContainer_WithEnvValueFrom(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := core.Container{
		Name: "test",
		Env: []core.EnvVar{
			{
				Name: "FIELD_REF",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "CONFIG_KEY",
				ValueFrom: &core.EnvVarSource{
					ConfigMapKeyRef: &core.ConfigMapKeySelector{
						LocalObjectReference: core.LocalObjectReference{Name: "config"},
						Key:                  "key1",
					},
				},
			},
			{
				Name: "SECRET_KEY",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{Name: "secret"},
						Key:                  "password",
					},
				},
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(len(dst.Env)).To(gomega.Equal(3))
	g.Expect(dst.Env[0].ValueFrom).ToNot(gomega.BeNil())
	g.Expect(dst.Env[0].ValueFrom.FieldRef).ToNot(gomega.BeNil())
	g.Expect(dst.Env[0].ValueFrom.FieldRef.FieldPath).To(gomega.Equal("metadata.name"))
	g.Expect(dst.Env[1].ValueFrom.ConfigMapKeyRef).ToNot(gomega.BeNil())
	g.Expect(dst.Env[1].ValueFrom.ConfigMapKeyRef.Key).To(gomega.Equal("key1"))
	g.Expect(dst.Env[2].ValueFrom.SecretKeyRef).ToNot(gomega.BeNil())
	g.Expect(dst.Env[2].ValueFrom.SecretKeyRef.Key).To(gomega.Equal("password"))
}

// TestConvertContainer_WithVolumeMounts tests convertContainer with VolumeMounts
func TestConvertContainer_WithVolumeMounts(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mountProp := core.MountPropagationBidirectional
	src := core.Container{
		Name: "test",
		VolumeMounts: []core.VolumeMount{
			{
				Name:             "data",
				MountPath:        "/data",
				SubPath:          "subdir",
				ReadOnly:         true,
				MountPropagation: &mountProp,
			},
		},
		VolumeDevices: []core.VolumeDevice{
			{
				Name:       "device1",
				DevicePath: "/dev/device1",
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(len(dst.VolumeMounts)).To(gomega.Equal(1))
	g.Expect(dst.VolumeMounts[0].Name).To(gomega.Equal("data"))
	g.Expect(dst.VolumeMounts[0].MountPath).To(gomega.Equal("/data"))
	g.Expect(dst.VolumeMounts[0].SubPath).To(gomega.Equal("subdir"))
	g.Expect(dst.VolumeMounts[0].ReadOnly).To(gomega.Equal(true))
	g.Expect(dst.VolumeMounts[0].MountPropagation).ToNot(gomega.BeNil())
	g.Expect(*dst.VolumeMounts[0].MountPropagation).To(gomega.Equal("Bidirectional"))
	g.Expect(len(dst.VolumeDevices)).To(gomega.Equal(1))
	g.Expect(dst.VolumeDevices[0].Name).To(gomega.Equal("device1"))
}

// TestConvertProbe tests the convertProbe function
func TestConvertProbe(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.Probe{
		ProbeHandler: core.ProbeHandler{
			Exec: &core.ExecAction{
				Command: []string{"cat", "/tmp/health"},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	dst := convertProbe(src)

	g.Expect(dst).ToNot(gomega.BeNil())
	g.Expect(dst.Exec).ToNot(gomega.BeNil())
	g.Expect(dst.Exec.Command).To(gomega.Equal([]string{"cat", "/tmp/health"}))
	g.Expect(dst.InitialDelaySeconds).To(gomega.Equal(int32(10)))
	g.Expect(dst.TimeoutSeconds).To(gomega.Equal(int32(5)))
	g.Expect(dst.PeriodSeconds).To(gomega.Equal(int32(30)))
	g.Expect(dst.SuccessThreshold).To(gomega.Equal(int32(1)))
	g.Expect(dst.FailureThreshold).To(gomega.Equal(int32(3)))
}

// TestConvertProbe_HTTPGet tests convertProbe with HTTPGet
func TestConvertProbe_HTTPGet(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.Probe{
		ProbeHandler: core.ProbeHandler{
			HTTPGet: &core.HTTPGetAction{
				Path:   "/health",
				Port:   intstr.FromInt(8080),
				Host:   "localhost",
				Scheme: core.URISchemeHTTPS,
				HTTPHeaders: []core.HTTPHeader{
					{Name: "X-Custom-Header", Value: "test"},
				},
			},
		},
	}

	dst := convertProbe(src)

	g.Expect(dst.HTTPGet).ToNot(gomega.BeNil())
	g.Expect(dst.HTTPGet.Path).To(gomega.Equal("/health"))
	g.Expect(dst.HTTPGet.Port).To(gomega.Equal(int32(8080)))
	g.Expect(dst.HTTPGet.Host).To(gomega.Equal("localhost"))
	g.Expect(dst.HTTPGet.Scheme).To(gomega.Equal("HTTPS"))
	g.Expect(len(dst.HTTPGet.HTTPHeaders)).To(gomega.Equal(1))
	g.Expect(dst.HTTPGet.HTTPHeaders[0].Name).To(gomega.Equal("X-Custom-Header"))
	g.Expect(dst.HTTPGet.HTTPHeaders[0].Value).To(gomega.Equal("test"))
}

// TestConvertProbe_TCPSocket tests convertProbe with TCPSocket
func TestConvertProbe_TCPSocket(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.Probe{
		ProbeHandler: core.ProbeHandler{
			TCPSocket: &core.TCPSocketAction{
				Port: intstr.FromInt(3306),
				Host: "db.example.com",
			},
		},
	}

	dst := convertProbe(src)

	g.Expect(dst.TCPSocket).ToNot(gomega.BeNil())
	g.Expect(dst.TCPSocket.Port).To(gomega.Equal(int32(3306)))
	g.Expect(dst.TCPSocket.Host).To(gomega.Equal("db.example.com"))
}

// TestConvertProbe_GRPC tests convertProbe with GRPC
func TestConvertProbe_GRPC(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	service := "my-service"
	src := &core.Probe{
		ProbeHandler: core.ProbeHandler{
			GRPC: &core.GRPCAction{
				Port:    9090,
				Service: &service,
			},
		},
	}

	dst := convertProbe(src)

	g.Expect(dst.GRPC).ToNot(gomega.BeNil())
	g.Expect(dst.GRPC.Port).To(gomega.Equal(int32(9090)))
	g.Expect(dst.GRPC.Service).ToNot(gomega.BeNil())
	g.Expect(*dst.GRPC.Service).To(gomega.Equal("my-service"))
}

// TestConvertContainer_WithProbes tests convertContainer with all probes
func TestConvertContainer_WithProbes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := core.Container{
		Name: "test",
		LivenessProbe: &core.Probe{
			ProbeHandler: core.ProbeHandler{
				Exec: &core.ExecAction{Command: []string{"liveness"}},
			},
		},
		ReadinessProbe: &core.Probe{
			ProbeHandler: core.ProbeHandler{
				HTTPGet: &core.HTTPGetAction{
					Path: "/ready",
					Port: intstr.FromInt(8080),
				},
			},
		},
		StartupProbe: &core.Probe{
			ProbeHandler: core.ProbeHandler{
				TCPSocket: &core.TCPSocketAction{
					Port: intstr.FromInt(8080),
				},
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(dst.LivenessProbe).ToNot(gomega.BeNil())
	g.Expect(dst.LivenessProbe.Exec).ToNot(gomega.BeNil())
	g.Expect(dst.ReadinessProbe).ToNot(gomega.BeNil())
	g.Expect(dst.ReadinessProbe.HTTPGet).ToNot(gomega.BeNil())
	g.Expect(dst.StartupProbe).ToNot(gomega.BeNil())
	g.Expect(dst.StartupProbe.TCPSocket).ToNot(gomega.BeNil())
}

// TestConvertLifecycleHandler tests the convertLifecycleHandler function
func TestConvertLifecycleHandler(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.LifecycleHandler{
		Exec: &core.ExecAction{
			Command: []string{"/bin/sh", "-c", "cleanup.sh"},
		},
	}

	dst := convertLifecycleHandler(src)

	g.Expect(dst).ToNot(gomega.BeNil())
	g.Expect(dst.Exec).ToNot(gomega.BeNil())
	g.Expect(dst.Exec.Command).To(gomega.Equal([]string{"/bin/sh", "-c", "cleanup.sh"}))
}

// TestConvertLifecycleHandler_HTTPGet tests convertLifecycleHandler with HTTPGet
func TestConvertLifecycleHandler_HTTPGet(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.LifecycleHandler{
		HTTPGet: &core.HTTPGetAction{
			Path: "/shutdown",
			Port: intstr.FromInt(8080),
		},
	}

	dst := convertLifecycleHandler(src)

	g.Expect(dst.HTTPGet).ToNot(gomega.BeNil())
	g.Expect(dst.HTTPGet.Path).To(gomega.Equal("/shutdown"))
	g.Expect(dst.HTTPGet.Port).To(gomega.Equal(int32(8080)))
}

// TestConvertContainer_WithLifecycle tests convertContainer with Lifecycle
func TestConvertContainer_WithLifecycle(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := core.Container{
		Name: "test",
		Lifecycle: &core.Lifecycle{
			PostStart: &core.LifecycleHandler{
				Exec: &core.ExecAction{Command: []string{"poststart.sh"}},
			},
			PreStop: &core.LifecycleHandler{
				HTTPGet: &core.HTTPGetAction{
					Path: "/prestop",
					Port: intstr.FromInt(8080),
				},
			},
		},
	}

	dst := convertContainer(src)

	g.Expect(dst.Lifecycle).ToNot(gomega.BeNil())
	g.Expect(dst.Lifecycle.PostStart).ToNot(gomega.BeNil())
	g.Expect(dst.Lifecycle.PostStart.Exec).ToNot(gomega.BeNil())
	g.Expect(dst.Lifecycle.PreStop).ToNot(gomega.BeNil())
	g.Expect(dst.Lifecycle.PreStop.HTTPGet).ToNot(gomega.BeNil())
}

// TestConvertSecurityContext tests the convertSecurityContext function
func TestConvertSecurityContext(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	privileged := true
	runAsUser := int64(1000)
	runAsGroup := int64(2000)
	runAsNonRoot := true
	readOnlyRootFS := true
	allowPrivEsc := false

	src := &core.SecurityContext{
		Privileged:               &privileged,
		RunAsUser:                &runAsUser,
		RunAsGroup:               &runAsGroup,
		RunAsNonRoot:             &runAsNonRoot,
		ReadOnlyRootFilesystem:   &readOnlyRootFS,
		AllowPrivilegeEscalation: &allowPrivEsc,
		Capabilities: &core.Capabilities{
			Add:  []core.Capability{"NET_ADMIN", "SYS_TIME"},
			Drop: []core.Capability{"ALL"},
		},
	}

	dst := convertSecurityContext(src)

	g.Expect(dst).ToNot(gomega.BeNil())
	g.Expect(dst.Privileged).ToNot(gomega.BeNil())
	g.Expect(*dst.Privileged).To(gomega.Equal(true))
	g.Expect(dst.RunAsUser).ToNot(gomega.BeNil())
	g.Expect(*dst.RunAsUser).To(gomega.Equal(int64(1000)))
	g.Expect(dst.RunAsGroup).ToNot(gomega.BeNil())
	g.Expect(*dst.RunAsGroup).To(gomega.Equal(int64(2000)))
	g.Expect(dst.RunAsNonRoute).ToNot(gomega.BeNil())
	g.Expect(*dst.RunAsNonRoute).To(gomega.Equal(true))
	g.Expect(dst.ReadOnlyRouteFilesystem).ToNot(gomega.BeNil())
	g.Expect(*dst.ReadOnlyRouteFilesystem).To(gomega.Equal(true))
	g.Expect(dst.AllowPrivilegeEscalation).ToNot(gomega.BeNil())
	g.Expect(*dst.AllowPrivilegeEscalation).To(gomega.Equal(false))
	g.Expect(dst.Capabilities).ToNot(gomega.BeNil())
	g.Expect(len(dst.Capabilities.Add)).To(gomega.Equal(2))
	g.Expect(dst.Capabilities.Add[0]).To(gomega.Equal("NET_ADMIN"))
	g.Expect(dst.Capabilities.Add[1]).To(gomega.Equal("SYS_TIME"))
	g.Expect(len(dst.Capabilities.Drop)).To(gomega.Equal(1))
	g.Expect(dst.Capabilities.Drop[0]).To(gomega.Equal("ALL"))
}

// TestConvertSecurityContext_WithSELinux tests convertSecurityContext with SELinux
func TestConvertSecurityContext_WithSELinux(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	src := &core.SecurityContext{
		SELinuxOptions: &core.SELinuxOptions{
			User:  "user1",
			Role:  "role1",
			Type:  "type1",
			Level: "s0:c123,c456",
		},
	}

	dst := convertSecurityContext(src)

	g.Expect(dst.SELinuxOptions).ToNot(gomega.BeNil())
	g.Expect(dst.SELinuxOptions.User).To(gomega.Equal("user1"))
	g.Expect(dst.SELinuxOptions.Role).To(gomega.Equal("role1"))
	g.Expect(dst.SELinuxOptions.Type).To(gomega.Equal("type1"))
	g.Expect(dst.SELinuxOptions.Level).To(gomega.Equal("s0:c123,c456"))
}

// TestConvertSecurityContext_WithSeccomp tests convertSecurityContext with Seccomp
func TestConvertSecurityContext_WithSeccomp(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	localhostProfile := "my-profile.json"
	src := &core.SecurityContext{
		SeccompProfile: &core.SeccompProfile{
			Type:             core.SeccompProfileTypeLocalhost,
			LocalhostProfile: &localhostProfile,
		},
	}

	dst := convertSecurityContext(src)

	g.Expect(dst.SeccompProfile).ToNot(gomega.BeNil())
	g.Expect(dst.SeccompProfile.Type).To(gomega.Equal("Localhost"))
	g.Expect(dst.SeccompProfile.LocalhostProfile).ToNot(gomega.BeNil())
	g.Expect(*dst.SeccompProfile.LocalhostProfile).To(gomega.Equal("my-profile.json"))
}

// TestConvertContainer_WithSecurityContext tests convertContainer with SecurityContext
func TestConvertContainer_WithSecurityContext(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	privileged := false
	src := core.Container{
		Name: "test",
		SecurityContext: &core.SecurityContext{
			Privileged: &privileged,
		},
	}

	dst := convertContainer(src)

	g.Expect(dst.SecurityContext).ToNot(gomega.BeNil())
	g.Expect(dst.SecurityContext.Privileged).ToNot(gomega.BeNil())
	g.Expect(*dst.SecurityContext.Privileged).To(gomega.Equal(false))
}

// TestAddon_With tests the Addon.With() method
func TestAddon_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	metadataJSON, _ := json.Marshal(metadata)

	addon := &crd.Addon{
		Spec: crd.AddonSpec{
			Container: core.Container{
				Name:    "addon-container",
				Image:   "addon-image:v1.0",
				Command: []string{"/bin/addon"},
				Args:    []string{"--verbose"},
			},
			Metadata: runtime.RawExtension{
				Raw: metadataJSON,
			},
		},
	}
	addon.Name = "test-addon"

	ext1 := crd.Extension{
		Spec: crd.ExtensionSpec{
			Addon: "test-addon",
			Container: core.Container{
				Name:  "ext1",
				Image: "ext1-image:v1.0",
			},
		},
	}
	ext1.Name = "extension-1"

	r := &Addon{}
	r.With(addon, ext1)

	g.Expect(r.Name).To(gomega.Equal("test-addon"))
	g.Expect(r.Container.Name).To(gomega.Equal("addon-container"))
	g.Expect(r.Container.Image).To(gomega.Equal("addon-image:v1.0"))
	g.Expect(r.Container.Command).To(gomega.Equal([]string{"/bin/addon"}))
	g.Expect(r.Container.Args).To(gomega.Equal([]string{"--verbose"}))
	g.Expect(len(r.Extensions)).To(gomega.Equal(1))
	g.Expect(r.Extensions[0].Name).To(gomega.Equal("extension-1"))
	g.Expect(r.Extensions[0].Addon).To(gomega.Equal("test-addon"))
}

// TestAddon_With_WithComplexContainer tests Addon.With() with complex container
func TestAddon_With_WithComplexContainer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	addon := &crd.Addon{
		Spec: crd.AddonSpec{
			Container: core.Container{
				Name:  "complex-addon",
				Image: "addon:latest",
				Ports: []core.ContainerPort{
					{
						Name:          "metrics",
						ContainerPort: 9090,
						Protocol:      core.ProtocolTCP,
					},
				},
				Env: []core.EnvVar{
					{Name: "LOG_LEVEL", Value: "debug"},
				},
				Resources: core.ResourceRequirements{
					Limits: core.ResourceList{
						core.ResourceMemory: resource.MustParse("512Mi"),
					},
					Requests: core.ResourceList{
						core.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
		},
	}
	addon.Name = "complex-addon"

	r := &Addon{}
	r.With(addon)

	g.Expect(r.Name).To(gomega.Equal("complex-addon"))
	g.Expect(r.Container.Name).To(gomega.Equal("complex-addon"))
	g.Expect(len(r.Container.Ports)).To(gomega.Equal(1))
	g.Expect(r.Container.Ports[0].Name).To(gomega.Equal("metrics"))
	g.Expect(r.Container.Ports[0].ContainerPort).To(gomega.Equal(int32(9090)))
	g.Expect(len(r.Container.Env)).To(gomega.Equal(1))
	g.Expect(r.Container.Env[0].Name).To(gomega.Equal("LOG_LEVEL"))
	g.Expect(r.Container.Env[0].Value).To(gomega.Equal("debug"))
	g.Expect(r.Container.Resources.Limits["memory"]).To(gomega.Equal("512Mi"))
	g.Expect(r.Container.Resources.Requests["memory"]).To(gomega.Equal("256Mi"))
}

// TestExtension_With tests the Extension.With() method
func TestExtension_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	metadata := map[string]interface{}{
		"priority": "high",
	}
	metadataJSON, _ := json.Marshal(metadata)

	ext := &crd.Extension{
		Spec: crd.ExtensionSpec{
			Addon: "parent-addon",
			Container: core.Container{
				Name:    "extension-container",
				Image:   "extension:v2.0",
				Command: []string{"/bin/extension"},
			},
			Metadata: runtime.RawExtension{
				Raw: metadataJSON,
			},
		},
	}
	ext.Name = "test-extension"

	r := &Extension{}
	r.With(ext)

	g.Expect(r.Name).To(gomega.Equal("test-extension"))
	g.Expect(r.Addon).To(gomega.Equal("parent-addon"))
	g.Expect(r.Container.Name).To(gomega.Equal("extension-container"))
	g.Expect(r.Container.Image).To(gomega.Equal("extension:v2.0"))
	g.Expect(r.Container.Command).To(gomega.Equal([]string{"/bin/extension"}))
}

// TestExtension_With_WithProbes tests Extension.With() with container probes
func TestExtension_With_WithProbes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ext := &crd.Extension{
		Spec: crd.ExtensionSpec{
			Addon: "parent",
			Container: core.Container{
				Name:  "ext-with-probes",
				Image: "extension:latest",
				LivenessProbe: &core.Probe{
					ProbeHandler: core.ProbeHandler{
						HTTPGet: &core.HTTPGetAction{
							Path: "/healthz",
							Port: intstr.FromInt(8080),
						},
					},
					InitialDelaySeconds: 30,
				},
				ReadinessProbe: &core.Probe{
					ProbeHandler: core.ProbeHandler{
						TCPSocket: &core.TCPSocketAction{
							Port: intstr.FromInt(8080),
						},
					},
					PeriodSeconds: 10,
				},
			},
		},
	}
	ext.Name = "probe-extension"

	r := &Extension{}
	r.With(ext)

	g.Expect(r.Container.LivenessProbe).ToNot(gomega.BeNil())
	g.Expect(r.Container.LivenessProbe.HTTPGet).ToNot(gomega.BeNil())
	g.Expect(r.Container.LivenessProbe.HTTPGet.Path).To(gomega.Equal("/healthz"))
	g.Expect(r.Container.LivenessProbe.InitialDelaySeconds).To(gomega.Equal(int32(30)))
	g.Expect(r.Container.ReadinessProbe).ToNot(gomega.BeNil())
	g.Expect(r.Container.ReadinessProbe.TCPSocket).ToNot(gomega.BeNil())
	g.Expect(r.Container.ReadinessProbe.PeriodSeconds).To(gomega.Equal(int32(10)))
}

// TestQuestion_With tests the Question.With() method
func TestQuestion_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Question{
		Order:       1,
		Text:        "What is your platform?",
		Explanation: "Select your target platform",
		IncludeFor: []model.CategorizedTag{
			{Category: "Language", Tag: "Java"},
		},
		ExcludeFor: []model.CategorizedTag{
			{Category: "Runtime", Tag: "Legacy"},
		},
		Answers: []model.Answer{
			{Order: 1, Text: "Cloud", Risk: "green"},
			{Order: 2, Text: "On-Premise", Risk: "yellow"},
		},
	}

	r := &Question{}
	r.With(m)

	g.Expect(r.Order).To(gomega.Equal(uint(1)))
	g.Expect(r.Text).To(gomega.Equal("What is your platform?"))
	g.Expect(r.Explanation).To(gomega.Equal("Select your target platform"))
	g.Expect(len(r.IncludeFor)).To(gomega.Equal(1))
	g.Expect(r.IncludeFor[0].Category).To(gomega.Equal("Language"))
	g.Expect(r.IncludeFor[0].Tag).To(gomega.Equal("Java"))
	g.Expect(len(r.ExcludeFor)).To(gomega.Equal(1))
	g.Expect(r.ExcludeFor[0].Category).To(gomega.Equal("Runtime"))
	g.Expect(len(r.Answers)).To(gomega.Equal(2))
	g.Expect(r.Answers[0].Text).To(gomega.Equal("Cloud"))
	g.Expect(r.Answers[1].Text).To(gomega.Equal("On-Premise"))
}

// TestQuestion_Model tests the Question.Model() method
func TestQuestion_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Question{
		Order:       1,
		Text:        "What is your platform?",
		Explanation: "Select your target platform",
		IncludeFor: []api.CategorizedTag{
			{Category: "Language", Tag: "Java"},
		},
		ExcludeFor: []api.CategorizedTag{
			{Category: "Runtime", Tag: "Legacy"},
		},
		Answers: []api.Answer{
			{Order: 1, Text: "Cloud", Risk: "green"},
			{Order: 2, Text: "On-Premise", Risk: "yellow"},
		},
	}

	m := r.Model()

	g.Expect(m.Order).To(gomega.Equal(uint(1)))
	g.Expect(m.Text).To(gomega.Equal("What is your platform?"))
	g.Expect(m.Explanation).To(gomega.Equal("Select your target platform"))
	g.Expect(len(m.IncludeFor)).To(gomega.Equal(1))
	g.Expect(m.IncludeFor[0].Category).To(gomega.Equal("Language"))
	g.Expect(len(m.ExcludeFor)).To(gomega.Equal(1))
	g.Expect(len(m.Answers)).To(gomega.Equal(2))
	g.Expect(m.Answers[0].Text).To(gomega.Equal("Cloud"))
}

// TestAnswer_With tests the Answer.With() method
func TestAnswer_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Answer{
		Order:        1,
		Text:         "Kubernetes",
		Risk:         "green",
		Rationale:    "Cloud native platform",
		Mitigation:   "None needed",
		Selected:     true,
		AutoAnswered: false,
		ApplyTags: []model.CategorizedTag{
			{Category: "Platform", Tag: "K8s"},
		},
		AutoAnswerFor: []model.CategorizedTag{
			{Category: "Container", Tag: "Docker"},
		},
	}

	r := &Answer{}
	r.With(m)

	g.Expect(r.Order).To(gomega.Equal(uint(1)))
	g.Expect(r.Text).To(gomega.Equal("Kubernetes"))
	g.Expect(r.Risk).To(gomega.Equal("green"))
	g.Expect(r.Rationale).To(gomega.Equal("Cloud native platform"))
	g.Expect(r.Mitigation).To(gomega.Equal("None needed"))
	g.Expect(r.Selected).To(gomega.Equal(true))
	g.Expect(r.AutoAnswered).To(gomega.Equal(false))
	g.Expect(len(r.ApplyTags)).To(gomega.Equal(1))
	g.Expect(r.ApplyTags[0].Category).To(gomega.Equal("Platform"))
	g.Expect(r.ApplyTags[0].Tag).To(gomega.Equal("K8s"))
	g.Expect(len(r.AutoAnswerFor)).To(gomega.Equal(1))
	g.Expect(r.AutoAnswerFor[0].Category).To(gomega.Equal("Container"))
}

// TestAnswer_Model tests the Answer.Model() method
func TestAnswer_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Answer{
		Order:        1,
		Text:         "Kubernetes",
		Risk:         "green",
		Rationale:    "Cloud native platform",
		Mitigation:   "None needed",
		Selected:     true,
		AutoAnswered: false,
		ApplyTags: []api.CategorizedTag{
			{Category: "Platform", Tag: "K8s"},
		},
		AutoAnswerFor: []api.CategorizedTag{
			{Category: "Container", Tag: "Docker"},
		},
	}

	m := r.Model()

	g.Expect(m.Order).To(gomega.Equal(uint(1)))
	g.Expect(m.Text).To(gomega.Equal("Kubernetes"))
	g.Expect(m.Risk).To(gomega.Equal("green"))
	g.Expect(m.Rationale).To(gomega.Equal("Cloud native platform"))
	g.Expect(m.Mitigation).To(gomega.Equal("None needed"))
	g.Expect(m.Selected).To(gomega.Equal(true))
	g.Expect(m.AutoAnswered).To(gomega.Equal(false))
	g.Expect(len(m.ApplyTags)).To(gomega.Equal(1))
	g.Expect(m.ApplyTags[0].Category).To(gomega.Equal("Platform"))
	g.Expect(len(m.AutoAnswerFor)).To(gomega.Equal(1))
}

// TestSection_With tests the Section.With() method
func TestSection_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.Section{
		Order:   1,
		Name:    "Application Details",
		Comment: "Provide application information",
		Questions: []model.Question{
			{Order: 1, Text: "What is the application type?"},
			{Order: 2, Text: "What framework is used?"},
		},
	}

	r := &Section{}
	r.With(m)

	g.Expect(r.Order).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("Application Details"))
	g.Expect(r.Comment).To(gomega.Equal("Provide application information"))
	g.Expect(len(r.Questions)).To(gomega.Equal(2))
	g.Expect(r.Questions[0].Order).To(gomega.Equal(uint(1)))
	g.Expect(r.Questions[0].Text).To(gomega.Equal("What is the application type?"))
	g.Expect(r.Questions[1].Order).To(gomega.Equal(uint(2)))
}

// TestSection_Model tests the Section.Model() method
func TestSection_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &Section{
		Order:   1,
		Name:    "Application Details",
		Comment: "Provide application information",
		Questions: []api.Question{
			{Order: 1, Text: "What is the application type?"},
			{Order: 2, Text: "What framework is used?"},
		},
	}

	m := r.Model()

	g.Expect(m.Order).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("Application Details"))
	g.Expect(m.Comment).To(gomega.Equal("Provide application information"))
	g.Expect(len(m.Questions)).To(gomega.Equal(2))
	g.Expect(m.Questions[0].Order).To(gomega.Equal(uint(1)))
	g.Expect(m.Questions[1].Order).To(gomega.Equal(uint(2)))
}

// TestBucket_With tests the Bucket.With() method
func TestBucket_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	now := time.Now()
	m := &model.Bucket{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Path:       "/tmp/bucket",
		Expiration: &now,
	}

	r := &Bucket{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Path).To(gomega.Equal("/tmp/bucket"))
	g.Expect(r.Expiration).ToNot(gomega.BeNil())
	g.Expect(*r.Expiration).To(gomega.Equal(now))
}

// TestImportSummary_With tests the ImportSummary.With() method
func TestImportSummary_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	now := time.Now()
	m := &model.ImportSummary{
		Model: model.Model{
			ID:         1,
			CreateTime: now,
		},
		Filename:       "import.csv",
		CreateEntities: true,
		Imports: []model.Import{
			{Processed: true, IsValid: true},
			{Processed: true, IsValid: true},
			{Processed: true, IsValid: false},
		},
	}

	r := &ImportSummary{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Filename).To(gomega.Equal("import.csv"))
	g.Expect(r.ImportTime).To(gomega.Equal(now))
	g.Expect(r.CreateEntities).To(gomega.Equal(true))
	g.Expect(r.ValidCount).To(gomega.Equal(2))
	g.Expect(r.InvalidCount).To(gomega.Equal(1))
	g.Expect(r.ImportStatus).To(gomega.Equal("Completed"))
}

// TestImportSummary_With_InProgress tests ImportSummary.With() with in-progress status
func TestImportSummary_With_InProgress(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m := &model.ImportSummary{
		Model: model.Model{ID: 1},
		Imports: []model.Import{
			{Processed: true, IsValid: true},
			{Processed: false, IsValid: false},
		},
	}

	r := &ImportSummary{}
	r.With(m)

	g.Expect(r.ValidCount).To(gomega.Equal(1))
	g.Expect(r.InvalidCount).To(gomega.Equal(0))
	g.Expect(r.ImportStatus).To(gomega.Equal("In Progress"))
}

// TestFile_With tests the File.With() method
func TestFile_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	now := time.Now()
	m := &model.File{
		Model: model.Model{
			ID:         1,
			CreateUser: "user1",
		},
		Name:       "document.pdf",
		Path:       "/files/document.pdf",
		Encoding:   "base64",
		Expiration: &now,
	}

	r := &File{}
	r.With(m)

	g.Expect(r.ID).To(gomega.Equal(uint(1)))
	g.Expect(r.Name).To(gomega.Equal("document.pdf"))
	g.Expect(r.Path).To(gomega.Equal("/files/document.pdf"))
	g.Expect(r.Encoding).To(gomega.Equal("base64"))
	g.Expect(r.Expiration).ToNot(gomega.BeNil())
	g.Expect(*r.Expiration).To(gomega.Equal(now))
}

// TestFile_Model tests the File.Model() method
func TestFile_Model(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	now := time.Now()
	r := &File{
		Resource:   Resource{ID: 1},
		Name:       "document.pdf",
		Path:       "/files/document.pdf",
		Encoding:   "base64",
		Expiration: &now,
	}

	m := r.Model()

	g.Expect(m.ID).To(gomega.Equal(uint(1)))
	g.Expect(m.Name).To(gomega.Equal("document.pdf"))
	g.Expect(m.Path).To(gomega.Equal("/files/document.pdf"))
	g.Expect(m.Encoding).To(gomega.Equal("base64"))
	g.Expect(m.Expiration).ToNot(gomega.BeNil())
	g.Expect(*m.Expiration).To(gomega.Equal(now))
}

// TestIdentityMap_With tests the IdentityMap.With() method
func TestIdentityMap_With(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	app := &Application{
		Identities: []IdentityRef{
			{ID: 1, Role: "source", Name: "identity1"},
			{ID: 2, Role: "maven", Name: "identity2"},
		},
	}

	r := make(IdentityMap)
	r.With(app)

	g.Expect(len(r)).To(gomega.Equal(2))
	_, exists1 := r[IdentityRef{ID: 1, Role: "source", Name: "identity1"}]
	g.Expect(exists1).To(gomega.Equal(true))
	_, exists2 := r[IdentityRef{ID: 2, Role: "maven", Name: "identity2"}]
	g.Expect(exists2).To(gomega.Equal(true))
}

// Application coordinates: model -> resource -> model round-trip.
func TestApplication_Coordinates_RoundTrip(t *testing.T) {
	g := gomega.NewWithT(t)

	// Build model with coordinates (model.Document via jsd helper).
	d := jsd.Document{Content: jsd.Map{"lang": "go"}, Schema: "http://schema/v1"}
	m := &model.Application{
		Model: model.Model{ID: 1},
		Name:  "app",
	}
	m.Coordinates = d.Model()

	// model -> resource
	r := &Application{}
	r.With(m, nil, nil)

	g.Expect(r.Coordinates).ToNot(gomega.BeNil())
	g.Expect(r.Coordinates.Schema).To(gomega.Equal("http://schema/v1"))
	g.Expect(r.Coordinates.Content["lang"]).To(gomega.Equal("go"))

	// resource -> model
	m2 := r.Model()
	g.Expect(m2.Coordinates).ToNot(gomega.BeNil())

	// Normalize back to jsd to check equality
	var round jsd.Document
	round.With(m2.Coordinates)
	g.Expect(round.Schema).To(gomega.Equal("http://schema/v1"))
	g.Expect(round.Content["lang"]).To(gomega.Equal("go"))
}

// AppTag helpers: WithRef and with()
func TestAppTag_WithRef_and_with(t *testing.T) {
	g := gomega.NewWithT(t)

	// WithRef should set Tag.ID and Source.
	ref := TagRef{ID: 42, Name: "tag42", Source: "manual"}
	at := &AppTag{}
	at.WithRef(&ref)
	g.Expect(at.Tag).ToNot(gomega.BeNil())
	g.Expect(at.Tag.ID).To(gomega.Equal(uint(42)))
	g.Expect(at.Source).To(gomega.Equal("manual"))

	// with() should copy TagID and Tag.
	mt := &model.ApplicationTag{
		ApplicationID: 7,
		TagID:         99,
		Source:        "auto",
		Tag:           model.Tag{Model: model.Model{ID: 99}, Name: "t99"},
	}
	at2 := &AppTag{}
	at2.with(mt)
	g.Expect(at2.ApplicationID).To(gomega.Equal(uint(7)))
	g.Expect(at2.TagID).To(gomega.Equal(uint(99)))
	g.Expect(at2.Tag).ToNot(gomega.BeNil())
	g.Expect(at2.Tag.ID).To(gomega.Equal(uint(99)))
	g.Expect(at2.Source).To(gomega.Equal("auto"))
}

// Task minimal: With/Patch and userPriority behavior.
func TestTask_With_and_Patch_and_userPriority(t *testing.T) {
	g := gomega.NewWithT(t)

	m := &model.Task{
		Model:    model.Model{ID: 1},
		Name:     "t",
		Kind:     "k",
		Addon:    "a",
		State:    "queued",
		Priority: 5, // userPriority should elevate when projected to resource
	}
	r := &Task{}
	r.With(m)

	g.Expect(r.Name).To(gomega.Equal("t"))
	g.Expect(r.Kind).To(gomega.Equal("k"))
	g.Expect(r.Addon).To(gomega.Equal("a"))
	g.Expect(r.Priority).To(gomega.Equal(15)) // elevated

	// Patch back
	m2 := &model.Task{}
	r.Patch(m2)
	g.Expect(m2.Name).To(gomega.Equal("t"))
	g.Expect(m2.Kind).To(gomega.Equal("k"))
	g.Expect(m2.Addon).To(gomega.Equal("a"))
	g.Expect(m2.Priority).To(gomega.Equal(15))
}

// FactKey: Qualify/Source/Name semantics.
func TestFactKey_Semantics(t *testing.T) {
	g := gomega.NewWithT(t)

	var k FactKey = "languages"
	k.Qualify("analysis")
	g.Expect(string(k)).To(gomega.Equal("analysis:languages"))

	source := k.Source()
	name := k.Name()
	g.Expect(source).To(gomega.Equal("analysis"))
	g.Expect(name).To(gomega.Equal("languages"))

	var onlySource FactKey = "analysis:"
	g.Expect(onlySource.Source()).To(gomega.Equal("analysis"))
	g.Expect(onlySource.Name()).To(gomega.Equal(""))
}
