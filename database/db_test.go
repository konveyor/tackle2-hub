package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"k8s.io/utils/env"
)

var N, _ = env.GetInt("TEST_CONCURRENT", 10)

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
	dq := make(chan int, N)
	for w := 0; w < N; w++ {
		go func(id int) {
			fmt.Printf("Started %d\n", id)
			for n := 0; n < N*10; n++ {
				m := &model.Setting{Key: fmt.Sprintf("key-%d-%d", id, n), Value: n}
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
