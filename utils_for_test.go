package sqlx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Stolen these codes from github.com/jmoiron/sqlx

func loadDefaultFixture(db *DB, t *testing.T) {
	tx := db.MustBeginTxm()
	tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Jason", "Moiron", "jmoiron@jmoiron.net")
	tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "John", "Doe", "johndoeDNE@gmail.net")
	tx.MustExec(tx.Rebind("INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)"), "United States", "New York", "1")
	tx.MustExec(tx.Rebind("INSERT INTO place (country, telcode) VALUES (?, ?)"), "Hong Kong", "852")
	tx.MustExec(tx.Rebind("INSERT INTO place (country, telcode) VALUES (?, ?)"), "Singapore", "65")
	if db.DriverName() == "mysql" {
		tx.MustExec(tx.Rebind("INSERT INTO capplace (`COUNTRY`, `TELCODE`) VALUES (?, ?)"), "Sarf Efrica", "27")
	} else {
		tx.MustExec(tx.Rebind("INSERT INTO capplace (\"COUNTRY\", \"TELCODE\") VALUES (?, ?)"), "Sarf Efrica", "27")
	}
	tx.MustExec(tx.Rebind("INSERT INTO employees (name, id) VALUES (?, ?)"), "Peter", "4444")
	tx.MustExec(tx.Rebind("INSERT INTO employees (name, id, boss_id) VALUES (?, ?, ?)"), "Joe", "1", "4444")
	tx.MustExec(tx.Rebind("INSERT INTO employees (name, id, boss_id) VALUES (?, ?, ?)"), "Martin", "2", "4444")
	tx.Commit()
}

func RunWithSchema(schema Schema, t *testing.T, test func(db *DB, t *testing.T)) {
	runner := func(db *DB, t *testing.T, create, drop string) {
		defer func() { MultiExec(db, drop) }()
		MultiExec(db, create)
		test(db, t)
	}

	if TestPostgres {
		create, drop := schema.Postgres()
		runner(pgdb, t, create, drop)
	}
	if TestSqlite {
		create, drop := schema.Sqlite3()
		runner(sldb, t, create, drop)
	}
	if TestMysql {
		create, drop := schema.MySQL()
		runner(mysqldb, t, create, drop)
	}
}

func MultiExec(e sqlx.Execer, query string) {
	stmts := strings.Split(query, ";\n")
	if len(strings.Trim(stmts[len(stmts)-1], " \n\t\r")) == 0 {
		stmts = stmts[:len(stmts)-1]
	}
	for _, s := range stmts {
		_, err := e.Exec(s)
		if err != nil {
			fmt.Println(err, s)
		}
	}
}
