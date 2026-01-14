package seed

import (
	"encoding/json"
	"os"

	"github.com/konveyor/tackle2-hub/internal/migration/v3/model"
	"gorm.io/gorm"
)

// RuleBundle seed object.
type RuleBundle struct {
	model.RuleBundle
	image    string
	excluded bool
}

// Create resources and files.
func (r *RuleBundle) Create(db *gorm.DB) {
	r.Image = &model.File{Name: "file.svg"}
	err := db.Create(r.Image).Error
	if err != nil {
		return
	}
	f, err := os.Create(r.Image.Path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.WriteString(r.image)
	if err != nil {
		return
	}
	_ = db.Create(&r.RuleBundle)
}

// Metadata builds windup metadata.
func Metadata(source, target string) (b []byte) {
	type MD struct {
		Source string `json:"source,omitempty"`
		Target string `json:"target,omitempty"`
	}
	b, _ = json.Marshal(&MD{
		Source: source,
		Target: target,
	})
	return
}

// Target builds metadata.
func Target(t string) (b []byte) {
	return Metadata("", t)
}
