package model

import (
	"gorm.io/gorm"
	"sync"
)

//
// depMutex ensures Dependency.Create() is not executed concurrently.
var depMutex sync.Mutex

type Dependency struct {
	Model
	ToID   uint         `gorm:"index"`
	To     *Application `gorm:"foreignKey:ToID;constraint:OnDelete:CASCADE"`
	FromID uint         `gorm:"index"`
	From   *Application `gorm:"foreignKey:FromID;constraint:OnDelete:CASCADE"`
}

//
// Create a dependency synchronized using a mutex.
func (r *Dependency) Create(db *gorm.DB) (err error) {
	depMutex.Lock()
	defer depMutex.Unlock()
	err = db.Create(r).Error
	return
}

//
// Validation Hook to avoid cyclic dependencies.
func (r *Dependency) BeforeCreate(db *gorm.DB) (err error) {
	var nextDeps []*Dependency
	var nextAppsIDs []uint
	nextAppsIDs = append(nextAppsIDs, r.FromID)
	for len(nextAppsIDs) != 0 {
		db.Where("ToID IN ?", nextAppsIDs).Find(&nextDeps)
		nextAppsIDs = nextAppsIDs[:0] // empty array, but keep capacity
		for _, nextDep := range nextDeps {
			if nextDep.FromID == r.ToID {
				err = DependencyCyclicError{}
				return
			}
			nextAppsIDs = append(nextAppsIDs, nextDep.FromID)
		}
	}

	return
}

//
// Custom error type to allow API recognize Cyclic Dependency error and assign proper status code.
type DependencyCyclicError struct{}

func (err DependencyCyclicError) Error() string {
	return "cyclic dependencies are not allowed"
}
