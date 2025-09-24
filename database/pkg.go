package database

import (
	"database/sql"
	"fmt"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	pg "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var log = logr.WithName("db")

var Settings = &settings.Settings

// Open the DB.
// prod = production (not migration).
func Open(prod bool) (db *gorm.DB, err error) {
	var driver gorm.Dialector
	if prod {
		dsn := "user=hub password=hub dbname=hub TimeZone=UTC"
		driver = pg.New(pg.Config{
			DSN:                  dsn,
			WithoutQuotingCheck:  true,
			PreferSimpleProtocol: !prod,
		})
		db, err = gorm.Open(
			driver,
			&gorm.Config{
				PrepareStmt:     prod,
				NamingStrategy:  &Namer{},
				CreateBatchSize: 500,
			})
	} else {
		Settings.DB.MaxConnection = 1
		dsn := fmt.Sprintf("file:%s?_journal=WAL", Settings.DB.Path)
		driver = sqlite.Open(dsn)
		db, err = gorm.Open(
			driver,
			&gorm.Config{
				PrepareStmt:     true,
				CreateBatchSize: 500,
				NamingStrategy: &schema.NamingStrategy{
					SingularTable: true,
					NoLowerCase:   true,
				},
			})
	}
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if Settings.DB.MaxConnection > 0 {
		dbx, nErr := db.DB()
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		dbx.SetMaxOpenConns(Settings.DB.MaxConnection)
	}
	err = db.AutoMigrate(model.PK{}, model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = PK.Load(db, []any{model.Setting{}})
	if err != nil {
		return
	}
	err = db.Callback().Create().Before("gorm:before_create").Register("assign-pk", assignPk)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Close the DB.
func Close(db *gorm.DB) (err error) {
	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

type Namer struct {
	schema.NamingStrategy
}

func (n Namer) TableName(table string) string {
	return strings.ToLower(table)
}

func (n Namer) ColumnName(_, column string) (name string) {
	name = strings.ToLower(column)
	if _, found := Keywords[name]; found {
		name = "\"" + name + "\""
	}
	return
}

var Keywords = map[string]struct{}{
	"all":               {},
	"analyse":           {},
	"analyze":           {},
	"and":               {},
	"any":               {},
	"array":             {},
	"as":                {},
	"asc":               {},
	"assertion":         {},
	"at":                {},
	"backward":          {},
	"begin":             {},
	"between":           {},
	"bigint":            {},
	"binary":            {},
	"bit":               {},
	"boolean":           {},
	"both":              {},
	"by":                {},
	"case":              {},
	"cast":              {},
	"char":              {},
	"character":         {},
	"character_varying": {},
	"check":             {},
	"collate":           {},
	"collation":         {},
	"column":            {},
	"constraint":        {},
	"constraints":       {},
	"create":            {},
	"cross":             {},
	"current_catalog":   {},
	"current_date":      {},
	"current_role":      {},
	"current_schema":    {},
	"current_time":      {},
	"current_timestamp": {},
	"current_user":      {},
	"cursor":            {},
	"cycle":             {},
	"date":              {},
	"datetime":          {},
	"default":           {},
	"deferrable":        {},
	"deferred":          {},
	"desc":              {},
	"distinct":          {},
	"do":                {},
	"domain":            {},
	"double":            {},
	"drop":              {},
	"else":              {},
	"end":               {},
	"except":            {},
	"false":             {},
	"fetch":             {},
	"for":               {},
	"foreign":           {},
	"from":              {},
	"full":              {},
	"grant":             {},
	"group":             {},
	"having":            {},
	"ilike":             {},
	"in":                {},
	"initially":         {},
	"inner":             {},
	"insert":            {},
	"int":               {},
	"integer":           {},
	"intersect":         {},
	"into":              {},
	"is":                {},
	"isnull":            {},
	"join":              {},
	"json":              {},
	"jsonb":             {},
	"lateral":           {},
	"left":              {},
	"like":              {},
	"limit":             {},
	"localtime":         {},
	"localtimestamp":    {},
	"natural":           {},
	"new":               {},
	"not":               {},
	"notnull":           {},
	"null":              {},
	"nullif":            {},
	"numeric":           {},
	"of":                {},
	"on":                {},
	"only":              {},
	"or":                {},
	"order":             {},
	"outer":             {},
	"over":              {},
	"overlaps":          {},
	"placing":           {},
	"primary":           {},
	"references":        {},
	"returning":         {},
	"right":             {},
	"select":            {},
	"session_user":      {},
	"set":               {},
	"similar":           {},
	"some":              {},
	"symmetric":         {},
	"table":             {},
	"tablesample":       {},
	"then":              {},
	"to":                {},
	"trailing":          {},
	"true":              {},
	"union":             {},
	"unique":            {},
	"user":              {},
	"using":             {},
	"when":              {},
	"where":             {},
	"window":            {},
	"with":              {},
}
