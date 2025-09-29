package postgres

import (
	"fmt"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Log      = logr.WithName("postgres")
	Settings = &settings.Settings
)

// Open the DB.
func Open(migration bool) (db *gorm.DB, err error) {
	var driver gorm.Dialector
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s TimeZone=UTC",
		Settings.DB.Host,
		Settings.DB.Port,
		Settings.DB.User,
		Settings.DB.Password,
		Settings.DB.Name)
	driver = pg.New(pg.Config{
		DSN:                  dsn,
		WithoutQuotingCheck:  true,
		PreferSimpleProtocol: migration,
	})
	db, err = open(
		driver,
		&gorm.Config{
			PrepareStmt:     !migration,
			CreateBatchSize: 500,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	pdb, nErr := db.DB()
	if nErr != nil {
		err = liberr.Wrap(nErr)
		return
	}
	pdb.SetMaxOpenConns(Settings.DB.MaxConnection)
	pdb.SetMaxIdleConns(Settings.DB.MaxConnection / 2)
	pdb.SetConnMaxLifetime(5 * time.Minute)
	err = db.AutoMigrate(model.PK{}, model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Callback().Create().Before("gorm:before_create").Register("assign-pk", func(db *gorm.DB) {

	})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// open used to open GORM with retries.
func open(driver gorm.Dialector, cfg *gorm.Config) (db *gorm.DB, err error) {
	for i := Settings.DB.Retries; i > 0; i-- {
		time.Sleep(time.Second)
		db, err = gorm.Open(driver, cfg)
		if err != nil {
			Log.Info(
				">> DATABASE CONNECTION FAILED <<",
				"reason",
				err.Error(),
				"retries",
				i)
		} else {
			break
		}
	}
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if Settings.DB.MaxConnection > 0 {
		pdb, nErr := db.DB()
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		pdb.SetMaxOpenConns(Settings.DB.MaxConnection)
		pdb.SetMaxIdleConns(Settings.DB.MaxConnection / 2)
		pdb.SetConnMaxLifetime(5 * time.Minute)
	}
	err = db.AutoMigrate(model.PK{}, model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
