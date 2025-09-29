package database

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/database/postgres"
	"gorm.io/gorm"
)

// Open the DB.
func Open(migration bool) (db *gorm.DB, err error) {
	db, err = postgres.Open(migration)
	return
}

// Close the DB.
func Close(db *gorm.DB) (err error) {
	pdb, err := db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = pdb.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
