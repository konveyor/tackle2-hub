package database

import (
	"encoding/json"
	"fmt"
	"github.com/konveyor/tackle2-hub/model"
	"os"
	"testing"
)

var N = 800

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
				uErr := db.Create(m).Error
				if uErr != nil {
					panic(uErr)
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
