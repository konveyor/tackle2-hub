package database

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/model"
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
			for n := 0; n < N; n++ {
				v, _ := json.Marshal(fmt.Sprintf("Test-%d", n))
				m := &model.Setting{Key: fmt.Sprintf("key-%d-%d", id, n), Value: v}
				fmt.Printf("(%.4d) CREATE: %.4d\n", id, n)
				uErr := db.Create(m).Error
				if uErr != nil {
					panic(uErr)
				}
				for i := 0; i < 10; i++ {
					tx := db.Begin()
					time.Sleep(time.Millisecond * time.Duration(n))
					uErr = tx.First(m).Error
					if uErr != nil {
						panic(uErr)
					}
					uErr = tx.Save(m).Error
					if uErr != nil {
						panic(uErr)
					}
					tx.Commit()
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
