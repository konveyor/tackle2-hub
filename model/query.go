package model

import (
	"gorm.io/gorm"
	"strings"
)

// Intersect returns an SQL intersect of the queries.
func Intersect(q ...*gorm.DB) (intersect *gorm.DB) {
	var part []string
	for n, q := range q {
		if n > 0 {
			part = append(part, "INTERSECT")
		}
		part = append(
			part,
			q.ToSQL(func(tx *gorm.DB) *gorm.DB {
				q = q.Session(&gorm.Session{DryRun: true})
				return q.Find(nil)
			}))
	}
	intersect = q[0].Raw(strings.Join(part, " "))
	return
}
