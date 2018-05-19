package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"

	sqlxx "github.com/jmoiron/sqlx"
)

// DB is a wrapper around *github.com/jmoiron/sqlx.DB which manages transaction.
type DB struct {
	*sqlxx.DB
	tx *Txm

	rollbacked *rollbacked
	activeTx   *activeTx
}

// Txm is a wrapper around *github.com/jmoiron/sqlx.DB with extra functionality and
// manages transaction.
type Txm struct {
	*sqlxx.Tx

	rollbacked *rollbacked
	activeTx   *activeTx
}

type activeTx struct{ count uint64 }
type rollbacked struct{ count uint64 }

// Open returns pointer of DB struct to manage transaction.
// It struct wrapped *github.com/jmoiron/sqlx.DB
// So we can use some methods of *github.com/jmoiron/sqlx.DB.
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sqlxx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db, activeTx: &activeTx{}}, err
}

// MustOpen returns only pointer of DB struct to manage transaction.
// But If you cause something error, It will do panic.
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

// SQL returns *sql.DB
// The reason for writing this method is that it needs to be
// written as *db.DB.DB to access *sql.DB.
func (db *DB) SQL() *sql.DB {
	return db.DB.DB
}

// setTx sets *github.com/jmoiron/sqlx.DB into *Txm.
func (db *DB) setTx(tx *sqlxx.Tx) {
	db.tx = &Txm{
		Tx:         tx,
		activeTx:   db.activeTx,
		rollbacked: &rollbacked{},
	}
}

// getTx gets *github.com/jmoiron/sqlx.DB into *Txm.
// It will increments count as activeTx.
func (db *DB) getTxm() *Txm {
	db.activeTx.increment()
	return db.tx
}

// BeginTxm begins a transaction and returns pointer of transaction manager.
// Actually, This method will invoke *github.com/jmoiron/sqlx.Beginx().
// but returns error if failed it.
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

// MustBeginTxm is like BeginTxm but panics
// if BeginTxm cannot begin transaction.
func (db *DB) MustBeginTxm() *Txm {
	txm, err := db.BeginTxm()
	if err != nil {
		panic(err)
	}
	return txm
}

// BeginTxmx begins a transaction and  returns pointer of transaction manager.
//
// The provided context is used until the transaction is committed or rolled
// back. If the context is canceled, the sql package will roll back the
// transaction. Tx.Commit will return an error if the context provided to
// BeginxContext is canceled.
func (db *DB) BeginTxmx(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
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

// MustBeginTxmx is like BeginTxmx but panics
// if BeginTxmx cannot begin transaction.
func (db *DB) MustBeginTxmx(ctx context.Context, opts *sql.TxOptions) (*Txm, error) {
	txm, err := db.BeginTxmx(ctx, opts)
	if err != nil {
		panic(err)
	}
	return txm, nil
}

// Commit commits the transaction.
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

// MustCommit is like Commit but panics if Commit is failed.
func (t *Txm) MustCommit() {
	defer t.reset()
	if err := t.Tx.Commit(); err != nil {
		panic(err)
	}
}

// Rollback rollbacks the transaction.
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

// MustRollback is like Rollback but panics if Rollback is failed.
func (t *Txm) MustRollback() {
	defer t.reset()
	if err := t.Tx.Rollback(); err != nil {
		panic(err)
	}
}

// reset resets some counter for transaction manager.
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
