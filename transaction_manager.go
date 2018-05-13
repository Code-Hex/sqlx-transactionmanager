package sqlx

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	sqlxx "github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlxx.DB
	mutex sync.Mutex
	pool  sync.Pool

	counter *activeTx
}

type Txm struct {
	*sqlxx.Tx
	mutex sync.Mutex

	counter *activeTx
}

type activeTx struct {
	count      uint64
	rollbacked uint64
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sqlxx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db, counter: &activeTx{}}, err
}

func (db *DB) Close() error { return db.DB.Close() }

func (db *DB) setTx(tx *sqlxx.Tx) {
	db.counter.incrementTx()
	db.pool.Put(&Txm{Tx: tx})
}

func (db *DB) getTxm() *Txm {
	return db.pool.Get().(*Txm)
}

func (db *DB) BeginTxm() (*Txm, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if !db.counter.HasActiveTx() {
		tx, err := db.DB.Beginx()
		if err != nil {
			return nil, err
		}
		db.setTx(tx)
		return db.getTxm(), nil
	}
	return db.getTxm(), nil
}

func (db *DB) MustBeginTxm() *Txm {
	txm, err := db.BeginTxm()
	if err != nil {
		panic(err)
	}
	return txm
}

func (db *DB) BeginTxxm(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if !db.counter.HasActiveTx() {
		tx, err := db.BeginTxx(ctx, opts)
		if err != nil {
			return nil, err
		}
		db.setTx(tx)
		return db.getTxm(), nil
	}
	return db.getTxm(), nil
}

func (db *DB) MustBeginTxxm(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
	txm, err := db.BeginTxxm(ctx, opts)
	if err != nil {
		panic(err)
	}
	return txm, nil
}

func (t *Txm) Commit() error {
	if t.counter.HasActiveTx() {
		return nil
	}
	return t.Tx.Commit()
}

func (t *Txm) Rollback() error {
	t.mutex.Lock()
	if t.counter.HasActiveTx() {
		t.counter.decrementTx()
		t.mutex.Unlock()
		return nil
	}
	t.mutex.Unlock()
	return t.Tx.Rollback()
}

func (a *activeTx) incrementTx() {
	atomic.AddUint64(&a.count, 1)
}

func (a *activeTx) decrementTx() {
	if a.HasActiveTx() {
		atomic.AddUint64(&a.count, ^uint64(0))
	}
}

func (a *activeTx) getActiveTx() uint64 {
	return atomic.LoadUint64(&a.count)
}

func (a *activeTx) HasActiveTx() bool {
	return a.getActiveTx() > 0
}
