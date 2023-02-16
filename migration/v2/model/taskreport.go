package model

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  JSON `gorm:"type:json"`
	Result    JSON `gorm:"type:json"`
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}
