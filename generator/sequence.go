package generator

import (
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// Sequence singleton generator.
var Sequence generator

// generator provides a primary key sequence.
type generator struct {
	keySet map[string]uint
	mutex  sync.Mutex
}

// Load highest key for all models.
func (r *generator) Load(db *gorm.DB, models []any) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.keySet = make(map[string]uint)
	for _, m := range models {
		mt := reflect.TypeOf(m)
		list := make([]map[string]any, 0)
		kind := strings.ToUpper(mt.Name())
		q := db.Table(kind)
		q = q.Select("ID")
		err = q.Find(&list).Error
		if err != nil {
			r.keySet[kind] = uint(0)
			continue
		}
		for _, m := range list {
			v := m["ID"]
			switch id := v.(type) {
			case uint:
				r.keySet[kind] = id
			case uint8:
				r.keySet[kind] = uint(id)
			case uint16:
				r.keySet[kind] = uint(id)
			case uint32:
				r.keySet[kind] = uint(id)
			case uint64:
				r.keySet[kind] = uint(id)
			case int:
				r.keySet[kind] = uint(id)
			case int8:
				r.keySet[kind] = uint(id)
			case int16:
				r.keySet[kind] = uint(id)
			case int32:
				r.keySet[kind] = uint(id)
			case int64:
				r.keySet[kind] = uint(id)
			}
		}
	}
	return
}

// Next returns the next primary key.
func (r *generator) Next(kind string) (id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	kind = strings.ToUpper(kind)
	id, found := r.keySet[kind]
	if found {
		id++
		r.keySet[kind] = id
	}
	return
}
