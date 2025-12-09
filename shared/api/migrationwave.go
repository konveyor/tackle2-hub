package api

import (
	"time"
)

// MigrationWave REST Resource
type MigrationWave struct {
	Resource          `yaml:",inline"`
	Name              string    `json:"name"`
	StartDate         time.Time `json:"startDate" yaml:"startDate" binding:"required"`
	EndDate           time.Time `json:"endDate" yaml:"endDate" binding:"required,gtfield=StartDate"`
	Applications      []Ref     `json:"applications"`
	Stakeholders      []Ref     `json:"stakeholders"`
	StakeholderGroups []Ref     `json:"stakeholderGroups" yaml:"stakeholderGroups"`
}
