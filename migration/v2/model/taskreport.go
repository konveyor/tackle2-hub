package model

type TaskReport struct {
	Model
	Status    string
	Error     string
	Total     int
	Completed int
	Activity  JSON
	Result    JSON
	TaskID    uint `gorm:"<-:create;uniqueIndex"`
	Task      *Task
}
