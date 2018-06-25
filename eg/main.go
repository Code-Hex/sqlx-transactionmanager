package main

import (
	"fmt"
	"os"
	"time"

	sqlx "github.com/Code-Hex/sqlx-transactionmanager"
	_ "github.com/go-sql-driver/mysql"
)

type Person struct {
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	AddedAt   time.Time `db:"added_at"`
}

func (p *Person) String() string {
	return fmt.Sprintf("%s %s: (%s) %s", p.FirstName, p.LastName, p.Email, p.AddedAt.String())
}

func dsn() string {
	// You can use environment vatiables from .envrc.
	// See https://github.com/direnv/direnv If you want to use .envrc.
	return os.Getenv("SQLX_MYSQL_DSN")
}

func loadDefaultFixture(db *sqlx.DB) {
	tx := db.MustBeginTxm()
	defer tx.Rollback()
	// If you want to know about tx.Rebind, See http://jmoiron.github.io/sqlx/#bindvars
	tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Jason", "Moiron", "jmoiron@jmoiron.net")
	tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "John", "Doe", "johndoeDNE@gmail.net")
	tx.Commit()
}

func Connect() *sqlx.DB {
	db := sqlx.MustOpen("mysql", dsn())
	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func main() {
	Mysql = true // use mysql
	db := Connect()
	defer db.Close()

	// See drivername
	fmt.Printf("Using: %s\n", db.DriverName())

	RunWithSchema(defaultSchema, db, DoTransaction(db))
}

// DoTransaction is example for transaction
// See transaction_manager_test.go if you want to know detail.
func DoTransaction(db *sqlx.DB) func(*sqlx.DB) {
	return func(db *sqlx.DB) {
		tx, err := db.BeginTxm()
		if err != nil {
			panic(err)
		}
		// Do rollbacks if fail something in transaction.
		// But do not commits if already commits in transaction.
		defer func() {
			if err := tx.Rollback(); err != nil {
				// Actually, you should do something...
				panic(err)
			}
		}()

		tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
		tx.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "a@b.com", "Code", "Hex")

		var p Person
		if err := tx.Get(&p, "SELECT * FROM person LIMIT 1"); err != nil {
			panic(err)
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}
		println(&p)
	}
}

func println(str fmt.Stringer) {
	fmt.Println(str)
}
