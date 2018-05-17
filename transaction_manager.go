package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"

	sqlxx "github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlxx.DB
	tx *Txm

	rollbacked *rollbacked
	activeTx   *activeTx
}

type Txm struct {
	*sqlxx.Tx

	rollbacked *rollbacked
	activeTx   *activeTx
}

type activeTx struct {
	count uint64
}

type rollbacked struct {
	count uint64
}

// Open returns pointer of DB struct to manage transaction
// It struct wrapped *github.com/jmoiron/sqlx.DB
// So we can use some methods of *github.com/jmoiron/sqlx.DB
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sqlxx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db, activeTx: &activeTx{}}, err
}

// MustOpen returns only pointer of DB struct to manage transaction
// But If you cause something error, It will do panic
func MustOpen(driverName, dataSourceName string) *DB {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

// Close closes *github.com/jmoiron/sqlx.DB
func (db *DB) Close() error {
	return db.DB.Close()
}

// Sql returns *sql.DB
// The reason for writing this method is that it needs to be written as db.DB.DB to access *sql.DB
func (db *DB) Sql() *sql.DB {
	return db.DB.DB
}

func (db *DB) setTx(tx *sqlxx.Tx) {
	db.tx = &Txm{
		Tx:         tx,
		activeTx:   db.activeTx,
		rollbacked: &rollbacked{},
	}
}

func (db *DB) getTxm() *Txm {
	db.activeTx.increment()
	return db.tx
}

// BeginTxm returns pointer of transaction manager
func (db *DB) BeginTxm() (*Txm, error) {
	if !db.activeTx.has() {
		tx, err := db.DB.Beginx()
		if err != nil {
			return nil, err
		}
		db.setTx(tx)
		return db.getTxm(), nil
	}
	return db.getTxm(), new(NestedBeginTxErr)
}

func (db *DB) MustBeginTxm() *Txm {
	txm, err := db.BeginTxm()
	if err != nil {
		panic(err)
	}
	return txm
}

func (db *DB) BeginTxxm(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
	if !db.activeTx.has() {
		tx, err := db.BeginTxx(ctx, opts)
		if err != nil {
			return nil, err
		}
		db.setTx(tx)
		return db.getTxm(), nil
	}
	return db.getTxm(), new(NestedBeginTxErr)
}

func (db *DB) MustBeginTxxm(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
	txm, err := db.BeginTxxm(ctx, opts)
	if err != nil {
		panic(err)
	}
	return txm, nil
}

func (t *Txm) Commit() error {
	if t.rollbacked.already() {
		return new(NestedCommitErr)
	}
	t.activeTx.decrement()
	if !t.activeTx.has() {
		return t.Tx.Commit()
	}
	return nil
}

func (t *Txm) Rollback() error {
	if !t.activeTx.has() {
		return nil
	}
	t.activeTx.decrement()
	if t.activeTx.has() {
		t.rollbacked.increment()
		return nil
	}
	return t.Tx.Rollback()
}

func (t *Txm) MustRollback() {
	defer t.reset()
	if err := t.Tx.Rollback(); err != nil {
		panic(err)
	}
}

func (t *Txm) MustCommit() {
	defer t.reset()
	if err := t.Tx.Commit(); err != nil {
		panic(err)
	}
}

func (t *Txm) reset() {
	t.rollbacked.reset()
	t.activeTx.reset()
}

func (r *rollbacked) String() string {
	return fmt.Sprintf("rollbacked in nested transaction: %d", r.times())
}

func (r *rollbacked) reset() {
	atomic.StoreUint64(&r.count, 0)
}

func (r *rollbacked) increment() {
	atomic.AddUint64(&r.count, 1)
}

func (r *rollbacked) times() uint64 {
	return atomic.LoadUint64(&r.count)
}

func (r *rollbacked) already() bool {
	return r.times() > 0
}

func (a *activeTx) String() string {
	return fmt.Sprintf("active tx counter: %d", a.get())
}

func (a *activeTx) reset() {
	atomic.StoreUint64(&a.count, 0)
}

func (a *activeTx) increment() {
	atomic.AddUint64(&a.count, 1)
}

func (a *activeTx) decrement() {
	if a.has() {
		atomic.AddUint64(&a.count, ^uint64(0))
	}
}

func (a *activeTx) get() uint64 {
	return atomic.LoadUint64(&a.count)
}

func (a *activeTx) has() bool {
	return a.get() > 0
}
