package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/internal/api"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	"k8s.io/utils/env"
)

var N, _ = env.GetInt("TEST_CONCURRENT", 10)

func TestDriver(t *testing.T) {
	pid := os.Getpid()
	Settings.DB.Path = fmt.Sprintf("/tmp/driver-%d.db", pid)
	defer func() {
		_ = os.Remove(Settings.DB.Path)
	}()
	db, err := Open(true)
	if err != nil {
		panic(err)
	}
	key := "driver"
	m := &model.Setting{Key: key, Value: "Test"}
	// insert.
	err = db.Create(m).Error
	if err != nil {
		panic(err)
	}
	// update
	err = db.Save(m).Error
	if err != nil {
		panic(err)
	}
	// select
	err = db.First(m, m.ID).Error
	if err != nil {
		panic(err)
	}
	// delete
	err = db.Delete(m).Error
	if err != nil {
		panic(err)
	}
}

func TestConcurrent(t *testing.T) {
	pid := os.Getpid()
	Settings.DB.Path = fmt.Sprintf("/tmp/concurrent-%d.db", pid)
	defer func() {
		_ = os.Remove(Settings.DB.Path)
	}()
	db, err := Open(true)
	if err != nil {
		panic(err)
	}

	type A struct {
		model.Model
	}

	type B struct {
		N int
		model.Model
		A   A
		AID uint
	}
	err = db.Migrator().AutoMigrate(&A{}, &B{})
	if err != nil {
		panic(err)
	}

	a := A{}
	err = db.Create(&a).Error
	if err != nil {
		panic(err)
	}

	dq := make(chan int, N)
	for w := 0; w < N; w++ {
		go func(id int) {
			fmt.Printf("Started %d\n", id)
			for n := 0; n < N; n++ {
				m := &B{N: n, A: a}
				m.CreateUser = "Test"
				fmt.Printf("(%.4d) CREATE: %.4d\n", id, n)
				uErr := db.Create(m).Error
				if uErr != nil {
					panic(uErr)
				}
				uErr = db.Save(m).Error
				if uErr != nil {
					panic(uErr)
				}
				for i := 0; i < 10; i++ {
					fmt.Printf("(%.4d) READ: %.4d/%.4d\n", id, n, i)
					uErr = db.First(m).Error
					if uErr != nil {
						panic(uErr)
					}
				}
				for i := 0; i < 10; i++ {
					fmt.Printf("(%.4d) LIST: %.4d/%.4d\n", id, n, i)
					page := api.Page{}
					cursor := api.Cursor{}
					mx := B{}
					dbx := db.Model(mx)
					dbx = dbx.Joins("A")
					dbx = dbx.Limit(10)
					cursor.With(dbx, page)
					for cursor.Next(&mx) {
						time.Sleep(time.Millisecond + 10)
						fmt.Printf("(%.4d) NEXT: %.4d/%.4d ID=%d\n", id, n, i, mx.ID)
					}
				}
				for i := 0; i < 4; i++ {
					uErr = db.Transaction(func(tx *gorm.DB) (err error) {
						time.Sleep(time.Millisecond * 10)
						for i := 0; i < 3; i++ {
							err = tx.Save(m).Error
							if err != nil {
								break
							}
						}
						return
					})
					if uErr != nil {
						panic(uErr)
					}
				}
			}
			dq <- id
		}(w)
	}
	for w := 0; w < N; w++ {
		id := <-dq
		fmt.Printf("Done %d\n", id)
	}
}

func TestKeyGen(t *testing.T) {
	pid := os.Getpid()
	Settings.DB.Path = fmt.Sprintf("/tmp/keygen-%d.db", pid)
	defer func() {
		_ = os.Remove(Settings.DB.Path)
	}()
	db, err := Open(true)
	if err != nil {
		panic(err)
	}
	// ids 1-7 created.
	N = 8
	for n := 1; n < N; n++ {
		m := &model.Setting{Key: fmt.Sprintf("key-%d", n), Value: n}
		err := db.Create(m).Error
		if err != nil {
			panic(err)
		}
		fmt.Printf("CREATED: %d/%d\n", m.ID, n)
		if uint(n) != m.ID {
			t.Errorf("id:%d but expected: %d", m.ID, n)
			return
		}
	}
	// delete ids=2,4,7.
	err = db.Delete(&model.Setting{}, []uint{2, 4, 7}).Error
	if err != nil {
		panic(err)
	}
	var count int64
	err = db.Model(&model.Setting{}).Where([]uint{2, 4, 7}).Count(&count).Error
	if err != nil {
		panic(err)
	}
	if count > 0 {
		t.Errorf("DELETED ids: 2,4,7 found.")
		return
	}
	// EXPLICIT id=10 created.
	m := &model.Setting{Key: "key-10", Value: 10}
	m.ID = 10
	err = db.Create(m).Error
	if err != nil {
		panic(err)
	}
	fmt.Printf("CREATED: %d\n", m.ID)
	if uint(10) != m.ID {
		t.Errorf("id:%d but expected: %d", m.ID, 10)
		return
	}
	// id=11 (next) created.
	next := uint(11)
	m = &model.Setting{Key: fmt.Sprintf("key-%d", next), Value: next}
	err = db.Create(m).Error
	if err != nil {
		panic(err)
	}
	fmt.Printf("CREATED: %d/%d (next)\n", m.ID, next)
	if next != m.ID {
		t.Errorf("id:%d but expected: %d", m.ID, next)
		return
	}
	// id =12,13 by list.
	list := []*model.Setting{
		{Key: "key-12", Value: 12},
		{Key: "key-13", Value: 13},
	}
	err = db.Create(&list).Error
	if err != nil {
		panic(err)
	}
	for id := range []uint{12, 13} {
		if next != m.ID {
			t.Errorf("id:%d but expected: %d", m.ID, id)
			return
		}
	}
	// EXPLICIT id=20,21 by list.
	list = []*model.Setting{
		{Model: model.Model{ID: 20}, Key: "key-20", Value: 20},
		{Model: model.Model{ID: 21}, Key: "key-21", Value: 21},
	}
	err = db.Create(&list).Error
	if err != nil {
		panic(err)
	}
	for id := range []uint{20, 21} {
		if next != m.ID {
			t.Errorf("id:%d but expected: %d", m.ID, id)
			return
		}
	}
	// EXPLICIT id=14 should not update the PK catalog.
	m = &model.Setting{Key: "key-14", Value: 14}
	m.ID = 14
	err = db.Save(m).Error
	if err != nil {
		panic(err)
	}
	// id=22 last one.
	m = &model.Setting{Key: "key-22", Value: 22}
	err = db.Save(m).Error
	if err != nil {
		panic(err)
	}
	if m.ID != 22 {
		t.Errorf("id:%d but expected: %d", m.ID, 22)
	}
}
