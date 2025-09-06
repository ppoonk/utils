package test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ppoonk/utils/instance"
)

// clear && go test -v instance/test/instance_test.go
func TestInstance(t *testing.T) {
	dbManager := instance.NewManager[*DBInstance]()

	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := dbManager.Add(fmt.Sprint(idx), newDBInstance(fmt.Sprint(idx)))
			if err != nil {
				t.Errorf("goroutine %d store DB failed: %v\n", idx, err)
				return
			}
			t.Logf("goroutine %d store DB\n", idx)
		}(i)
	}
	wg.Wait()

	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			db, err := dbManager.Load(fmt.Sprint(idx))
			if err != nil {
				t.Errorf("goroutine %d get DB failed: %v\n", idx, err)
				return
			}
			t.Logf("goroutine %d get DB: %p\n", idx, db)
		}(i)
	}
	wg.Wait()

}

type DBInstance struct {
	name string
}

func (db *DBInstance) Destroy() error {
	return nil
}

func newDBInstance(name string) *DBInstance {
	return &DBInstance{name: name}
}
