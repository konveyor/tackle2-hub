package reflect

import (
	"gorm.io/gorm"
)

// Select returns DB.Select() with validated fields.
func Select(in *gorm.DB, m any, fields ...string) (out *gorm.DB) {
	fields, err := HasFields(m, fields...)
	out = in.Select(fields)
	if err != nil {
		out.Statement.Error = err
		return
	}
	return
}

// Omit returns DB.Omit() with validated fields.
func Omit(in *gorm.DB, m any, fields ...string) (out *gorm.DB) {
	fields, err := HasFields(m, fields...)
	out = in.Omit(fields...)
	if err != nil {
		out.Statement.Error = err
		return
	}
	return
}
