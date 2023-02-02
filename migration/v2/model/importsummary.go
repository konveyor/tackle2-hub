package model

type ImportSummary struct {
	Model
	Content        []byte
	Filename       string
	ImportStatus   string
	Imports        []Import `gorm:"constraint:OnDelete:CASCADE"`
	CreateEntities bool
}
