package model

import (
	"testing"

	"github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TestDependencyBeforeCreate tests cyclic dependency detection.
func TestDependencyBeforeCreate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup in-memory database
	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	g.Expect(err).To(gomega.BeNil())

	// Auto-migrate the tables
	err = db.AutoMigrate(&Application{}, &Dependency{})
	g.Expect(err).To(gomega.BeNil())

	// Create test applications
	app1 := &Application{Name: "App1"}
	app2 := &Application{Name: "App2"}
	app3 := &Application{Name: "App3"}
	app4 := &Application{Name: "App4"}

	err = db.Create(app1).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(app2).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(app3).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(app4).Error
	g.Expect(err).To(gomega.BeNil())

	// Test 1: Valid dependency creation
	dep1 := &Dependency{FromID: app1.ID, ToID: app2.ID}
	err = dep1.Create(db)
	g.Expect(err).To(gomega.BeNil())

	// Test 2: Another valid dependency (chain)
	dep2 := &Dependency{FromID: app2.ID, ToID: app3.ID}
	err = dep2.Create(db)
	g.Expect(err).To(gomega.BeNil())

	// Test 3: Direct cycle should fail (A→B, B→A)
	directCycle := &Dependency{FromID: app2.ID, ToID: app1.ID}
	err = directCycle.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok := err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())

	// Test 4: Indirect cycle should fail (A→B→C, C→A)
	indirectCycle := &Dependency{FromID: app3.ID, ToID: app1.ID}
	err = indirectCycle.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok = err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())

	// Test 5: Valid parallel path (A→B→C and A→C)
	parallelPath := &Dependency{FromID: app1.ID, ToID: app3.ID}
	err = parallelPath.Create(db)
	g.Expect(err).To(gomega.BeNil())

	// Test 6: Another direct cycle should fail (B→C, C→B)
	anotherDirectCycle := &Dependency{FromID: app3.ID, ToID: app2.ID}
	err = anotherDirectCycle.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok = err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())

	// Test 7: Valid dependency from unconnected app
	dep3 := &Dependency{FromID: app4.ID, ToID: app1.ID}
	err = dep3.Create(db)
	g.Expect(err).To(gomega.BeNil())

	// Test 8: Cycle through longer chain should fail (D→A→B→C, C→D)
	longerCycle := &Dependency{FromID: app3.ID, ToID: app4.ID}
	err = longerCycle.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok = err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())
}

// TestDependencyBeforeCreate_SelfLoop tests self-referencing dependency.
func TestDependencyBeforeCreate_SelfLoop(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup in-memory database
	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	g.Expect(err).To(gomega.BeNil())

	// Auto-migrate the tables
	err = db.AutoMigrate(&Application{}, &Dependency{})
	g.Expect(err).To(gomega.BeNil())

	// Create test application
	app := &Application{Name: "App"}
	err = db.Create(app).Error
	g.Expect(err).To(gomega.BeNil())

	// Test: Self-loop should fail (A→A)
	selfLoop := &Dependency{FromID: app.ID, ToID: app.ID}
	err = selfLoop.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok := err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())
}

// TestDependencyBeforeCreate_ComplexGraph tests visited tracking with diamond pattern.
func TestDependencyBeforeCreate_ComplexGraph(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup in-memory database
	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	g.Expect(err).To(gomega.BeNil())

	// Auto-migrate the tables
	err = db.AutoMigrate(&Application{}, &Dependency{})
	g.Expect(err).To(gomega.BeNil())

	// Create test applications for diamond pattern:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	appA := &Application{Name: "AppA"}
	appB := &Application{Name: "AppB"}
	appC := &Application{Name: "AppC"}
	appD := &Application{Name: "AppD"}

	err = db.Create(appA).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(appB).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(appC).Error
	g.Expect(err).To(gomega.BeNil())
	err = db.Create(appD).Error
	g.Expect(err).To(gomega.BeNil())

	// Build diamond: A→B, A→C, B→D, C→D
	err = (&Dependency{FromID: appA.ID, ToID: appB.ID}).Create(db)
	g.Expect(err).To(gomega.BeNil())

	err = (&Dependency{FromID: appA.ID, ToID: appC.ID}).Create(db)
	g.Expect(err).To(gomega.BeNil())

	err = (&Dependency{FromID: appB.ID, ToID: appD.ID}).Create(db)
	g.Expect(err).To(gomega.BeNil())

	err = (&Dependency{FromID: appC.ID, ToID: appD.ID}).Create(db)
	g.Expect(err).To(gomega.BeNil())

	// Test: D→A should fail (creates cycle through multiple paths)
	cycle := &Dependency{FromID: appD.ID, ToID: appA.ID}
	err = cycle.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok := err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())

	// Test: D→B should fail (creates cycle)
	cycle2 := &Dependency{FromID: appD.ID, ToID: appB.ID}
	err = cycle2.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok = err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())

	// Test: D→C should fail (creates cycle)
	cycle3 := &Dependency{FromID: appD.ID, ToID: appC.ID}
	err = cycle3.Create(db)
	g.Expect(err).NotTo(gomega.BeNil())
	_, ok = err.(DependencyCyclicError)
	g.Expect(ok).To(gomega.BeTrue())
}
