package seed

import (
	"errors"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// Generator applies Generator seeds.
type Generator struct {
	generators []libseed.Generator
}

// With collects all the Generator seeds.
func (r *Generator) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		r.generators = append(r.generators, item.(libseed.Generator))
	}
	return
}

// Apply seeds the database with Generators.
func (r *Generator) Apply(db *gorm.DB) (err error) {
	log.Info("Applying Generators", "count", len(r.generators))
	for i := range r.generators {
		g := r.generators[i]
		m, found, fErr := r.find(db, "uuid = ?", g.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && m.Name != g.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", g.Name, m.ID)
			if fErr != nil {
				err = fErr
				return
			}
			if collides {
				err = r.rename(db, collision)
				if err != nil {
					return
				}
			}
		} else if !found {
			m, found, fErr = r.find(db, "name = ?", g.Name)
			if fErr != nil {
				err = fErr
				return
			}
			if found && m.CreateUser != "" {
				err = r.rename(db, m)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				m = &model.Generator{}
			}
		}
		m.UUID = &g.UUID
		m.Kind = g.Kind
		m.Name = g.Name
		m.Description = g.Description
		m.Values = g.Values
		m.Params = g.Params
		m.Repository = model.Repository(g.Repository)
		err = db.Save(m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// Convenience method to find a Generator.
func (r *Generator) find(db *gorm.DB, conditions ...any) (m *model.Generator, found bool, err error) {
	m = &model.Generator{}
	err = db.First(m, conditions...).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = liberr.Wrap(err)
		return
	}
	found = true
	return
}

// Rename a Generator by adding a suffix.
func (r *Generator) rename(db *gorm.DB, m *model.Generator) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", m.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			m.Name = newName
			err = db.Save(m).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			return
		}
	}
}
