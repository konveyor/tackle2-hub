package model

import "time"

type Tracker struct {
	Model
	Name        string `gorm:"index;unique;not null"`
	URL         string
	Kind        string
	Identity    *Identity
	IdentityID  uint
	Metadata    JSON
	Connected   bool
	LastUpdated time.Time
	Message     string
	Tickets     []Ticket `gorm:"constraint:OnDelete:CASCADE"`
}
