package tm

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// SQL interface implements for *sql.DB or wrapped it.
type SQL interface{ Begin() (*sql.Tx, error) }

// SQLx interface implements for *sqlx.DB or wrapped it.
type SQLx interface{ Beginx() (*sqlx.Tx, error) }

// Executor interface implements for *sql.Tx or wrapped it.
// It has'nt Commit and Rollback methods.
type Executor interface {
	Exec(string, ...interface{}) (sql.Result, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Stmt(*sql.Stmt) *sql.Stmt
	StmtContext(context.Context, *sql.Stmt) *sql.Stmt
}

// Executorx interface implements for *sqlx.Tx or wrapped it.
// It has'nt Commit and Rollback methods.
type Executorx interface {
	Executor

	Get(interface{}, string, ...interface{}) error
	GetContext(context.Context, interface{}, string, ...interface{}) error
	MustExec(string, ...interface{}) sql.Result
	MustExecContext(context.Context, string, ...interface{}) sql.Result
	NamedExec(string, interface{}) (sql.Result, error)
	NamedExecContext(context.Context, string, interface{}) (sql.Result, error)
	NamedQuery(string, interface{}) (*sqlx.Rows, error)
	NamedStmt(stmt *sqlx.NamedStmt) *sqlx.NamedStmt
	NamedStmtContext(context.Context, *sqlx.NamedStmt) *sqlx.NamedStmt
	PrepareNamedContext(context.Context, string) (*sqlx.NamedStmt, error)
	Preparex(string) (*sqlx.Stmt, error)
	PreparexContext(context.Context, string) (*sqlx.Stmt, error)
	QueryRowx(string, ...interface{}) *sqlx.Row
	QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row
	Queryx(string, ...interface{}) (*sqlx.Rows, error)
	QueryxContext(context.Context, string, ...interface{}) (*sqlx.Rows, error)
	Rebind(string) string
	Select(interface{}, string, ...interface{}) error
	SelectContext(context.Context, interface{}, string, ...interface{}) error
	Stmtx(interface{}) *sqlx.Stmt
	StmtxContext(context.Context, interface{}) *sqlx.Stmt
	Unsafe() *sqlx.Tx
}

// TxnFunc implemtnts for func(Executor) error
type TxnFunc func(Executor) error

// TxnxFunc implemtnts for func(Executorx) error
type TxnxFunc func(Executorx) error

// Run begins transaction around TxnFunc.
// It returns error and rollbacks if TxnFunc is failed.
// It commits if TxnFunc is successed.
func Run(db SQL, f TxnFunc) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// Runx begins transaction around TxnxFunc.
// It returns error and rollbacks if TxnxFunc is failed.
// It commits if TxnxFunc is successed.
func Runx(db SQLx, f TxnxFunc) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
