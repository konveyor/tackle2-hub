package model

//
// Target - analysis rule selector.
type Target struct {
	Model
	UUID        *string `gorm:"uniqueIndex"`
	Name        string  `gorm:"uniqueIndex;not null"`
	Description string
	Choice      bool
	Labels      JSON `gorm:"type:json"`
	ImageID     uint `gorm:"index" ref:"file"`
	Image       *File
	RuleSetID   *uint `gorm:"index"`
	RuleSet     *RuleSet
}

func (r *Target) Builtin() bool {
	return r.UUID != nil
}
