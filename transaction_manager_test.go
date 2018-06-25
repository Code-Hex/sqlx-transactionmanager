package sqlx

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	sqlxx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var (
	TestPostgres = true
	TestSqlite   = true
	TestMysql    = true
)

var sldb *DB
var pgdb *DB
var mysqldb *DB
var active = []*sqlxx.DB{}

const indent = "    "

func init() {
	ConnectAll()
}

func ConnectAll() {
	pgdsn := os.Getenv("SQLX_POSTGRES_DSN")
	mydsn := os.Getenv("SQLX_MYSQL_DSN")
	sqdsn := os.Getenv("SQLX_SQLITE_DSN")

	TestPostgres = pgdsn != "skip"
	TestMysql = mydsn != "skip"
	TestSqlite = sqdsn != "skip"

	if !strings.Contains(mydsn, "parseTime=true") {
		mydsn += "?parseTime=true"
	}

	if TestPostgres {
		pgdb = MustOpen("postgres", pgdsn)
	} else {
		fmt.Println("Disabling Postgres tests")
	}

	if TestMysql {
		mysqldb = MustOpen("mysql", mydsn)
	} else {
		fmt.Println("Disabling MySQL tests")
	}

	if TestSqlite {
		sldb = MustOpen("sqlite3", sqdsn)
	} else {
		fmt.Println("Disabling SQLite tests")
	}
}

type Schema struct {
	create string
	drop   string
}

func (s Schema) Postgres() (string, string) {
	return s.create, s.drop
}

func (s Schema) MySQL() (string, string) {
	return strings.Replace(s.create, `"`, "`", -1), s.drop
}

func (s Schema) Sqlite3() (string, string) {
	return strings.Replace(s.create, `now()`, `CURRENT_TIMESTAMP`, -1), s.drop
}

var defaultSchema = Schema{
	create: `
CREATE TABLE person (
	first_name text,
	last_name text,
	email text,
	added_at timestamp default now()
);

CREATE TABLE place (
	country text,
	city text NULL,
	telcode integer
);

`,
	drop: `
drop table person;
drop table place;
`,
}

type Person struct {
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	AddedAt   time.Time `db:"added_at"`
}

type Place struct {
	Country string
	City    sql.NullString
	TelCode int
}

func TestCommit(t *testing.T) {
	RunWithSchema(defaultSchema, t, func(db *DB, t *testing.T) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
		tx.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "a@b.com", "Code", "Hex")
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		var author Person
		if err := db.Get(&author, "SELECT * FROM person LIMIT 1"); err != nil {
			t.Fatal(
				errors.Wrapf(err, "commit test is failed\n    %s\n    %s\n",
					fmt.Sprintf("rollbacked in nested transaction: %d", db.rollbacked.times()),
					fmt.Sprintf("active tx counter: %d", db.activeTx.get()),
				),
			)

		}
		if author.FirstName != "Code" || author.LastName != "Hex" {
			t.Fatal("Failed to test commit")
		}

		tx2, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		tx2.MustExec(tx.Rebind("DELETE FROM person"))
		tx2.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Al", "paca", "kei@gmail.com")
		tx2.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "c@d.com", "Al", "paca")
		if err := tx2.Commit(); err != nil {
			t.Fatal(err)
		}

		var author2 Person
		if err := db.Get(&author2, "SELECT * FROM person LIMIT 1"); err != nil {
			t.Fatal(
				errors.Wrapf(err, "%s\n%s\n",
					fmt.Sprintf("rollbacked in nested transaction: %d", db.rollbacked.times()),
					fmt.Sprintf("active tx counter: %d", db.activeTx.get()),
				),
			)
		}
		if author2.FirstName != "Al" || author2.LastName != "paca" {
			t.Fatal("Failed to test commit2")
		}
		fmt.Fprintf(os.Stderr, "rollbacked in nested transaction: %d\n", db.rollbacked.times())
		fmt.Fprintf(os.Stderr, "active tx counter: %d\n", db.activeTx.get())
	})
}

func TestRollback(t *testing.T) {
	RunWithSchema(defaultSchema, t, func(db *DB, t *testing.T) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
		tx.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "a@b.com", "Code", "Hex")
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}

		var author Person
		if err := db.Get(&author, "SELECT * FROM person LIMIT 1"); err != sql.ErrNoRows {
			t.Fatal(
				errors.Wrapf(err, "rollback test is failed\n    %s\n    %s\n",
					fmt.Sprintf("rollbacked in nested transaction: %d", db.rollbacked.times()),
					fmt.Sprintf("active tx counter: %d", db.activeTx.get()),
				),
			)
		}
	})
}

func TestNestedCommit(t *testing.T) {
	nested := func(db *DB) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		if tx == nil {
			t.Fatal("Failed to return tx")
		}
		if !tx.activeTx.has() {
			t.Fatal("Failed having active transaction in nested BEGIN")
		}
	}
	RunWithSchema(defaultSchema, t, func(db *DB, t *testing.T) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
		tx.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "a@b.com", "Code", "Hex")

		// I will try begin 4 times
		nested(db)
		nested(db)
		nestedmore := func(db *DB) {
			tx, err := db.BeginTxm()
			if err != nil {
				t.Fatal(err)
			}
			nested(db)
			if tx == nil {
				t.Fatal("Failed to return tx")
			}
			if !tx.activeTx.has() {
				t.Fatal("Failed having active transaction in nested BEGIN")
			}
		}
		nestedmore(db)

		// Original begin + 4 times of nested begin
		for i := 0; i < 5; i++ {
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		}
		var author Person
		if err := db.Get(&author, "SELECT * FROM person LIMIT 1"); err != nil {
			t.Fatal(
				errors.Wrapf(err, "nested transaction test is failed\n    %s\n    %s\n",
					fmt.Sprintf("rollbacked in nested transaction: %d", db.rollbacked.times()),
					fmt.Sprintf("active tx counter: %d", db.activeTx.get()),
				),
			)
		}
		if err := tx.Commit(); err != sql.ErrTxDone {
			t.Fatal("Failed to cause error for already committed")
		}
	})
}

func TestNestedRollback(t *testing.T) {
	nested := func(db *DB) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.MustRollback()
		tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
		if !tx.activeTx.has() {
			t.Fatal("Failed having active transaction in nested BEGIN")
		}
		panic("Something failed")
		// Maybe we will `tx.Commit()` at last
	}
	nestedmore := func(db *DB) {
		tx, err := db.BeginTxm()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.MustRollback()
		nested(db)
		tx.Commit()
	}
	RunWithSchema(defaultSchema, t, func(db *DB, t *testing.T) {
		func() {
			tx, err := db.BeginTxm()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.MustRollback()
			tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
			nestedmore(db)
			tx.Commit()
		}()

		var author Person
		if err := db.Get(&author, "SELECT * FROM person LIMIT 1"); err != sql.ErrNoRows {
			t.Fatal(
				errors.Wrapf(err, "rollback test is failed\n    %s\n    %s\n",
					fmt.Sprintf("rollbacked in nested transaction: %d", db.rollbacked.times()),
					fmt.Sprintf("active tx counter: %d", db.activeTx.get()),
				),
			)
		}
	})
}
