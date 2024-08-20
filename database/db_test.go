package database

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/model"
	"k8s.io/utils/env"
)

var N, _ = env.GetInt("TEST_CONCURRENT", 10)

func TestConcurrent(t *testing.T) {
	Settings.DB.Path = "/tmp/concurrent.db"
	_ = os.Remove(Settings.DB.Path)
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
					fmt.Printf("(%.4d) BEGIN: %.4d/%.4d\n", id, n, i)
					tx := db.Begin()
					fmt.Printf("(%.4d) FIRST: %.4d/%.4d\n", id, n, i)
					uErr = tx.First(m).Error
					if uErr != nil {
						panic(uErr)
					}
					fmt.Printf("(%.4d) SAVE: %.4d/%.4d\n", id, n, i)
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
