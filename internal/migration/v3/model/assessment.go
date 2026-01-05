package model

type Review struct {
	Model
	BusinessCriticality uint   `gorm:"not null"`
	EffortEstimate      string `gorm:"not null"`
	ProposedAction      string `gorm:"not null"`
	WorkPriority        uint   `gorm:"not null"`
	Comments            string
	ApplicationID       uint `gorm:"uniqueIndex"`
	Application         *Application
}
