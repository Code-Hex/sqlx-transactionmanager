package sqlx

import (
	"sync"
	"testing"
)

func TestAtomicCount(t *testing.T) {
	RunWithSchema(defaultSchema, t, func(db *DB, t *testing.T) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		times := 1000000
		for i := 1; i < times; i++ {
			wg.Add(1)
			go func(d *DB) {
				defer wg.Done()
				_, err := db.BeginTxm()
				if e, ok := err.(*NestedBeginTxErr); !ok {
					panic(e)
				}
			}(db)
		}
		wg.Wait()

		if uint64(times) != db.activeTx.get() {
			t.Fatalf("Failed to atomic count in db activeTx: %d, expected %d", db.activeTx.get(), times)
		}
		if uint64(times) != tx.activeTx.get() {
			t.Fatalf("Failed to atomic count in tx activeTx: %d, expected %d", tx.activeTx.get(), times)
		}

		for i := 1; i < times; i++ {
			wg.Add(1)
			go func(txm *Txm) {
				defer wg.Done()
				if err := txm.Rollback(); err != nil {
					panic(err)
				}
			}(tx)
		}
		wg.Wait()

		// If rollback is failed, another test will fail after this test
		if err := tx.Rollback(); tx.activeTx.has() || err != nil {
			t.Fatalf("Failed to many rollback: error(%s), activeTx(%d)", err, tx.activeTx.get())
		}
	})
}
