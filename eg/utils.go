package main

import (
	"fmt"
	"strings"

	sqlx "github.com/Code-Hex/sqlx-transactionmanager"
	osqlx "github.com/jmoiron/sqlx"
)

var (
	Postgres bool
	Mysql    bool
	Sqlite   bool
)

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

func MultiExec(e osqlx.Execer, query string) {
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

func RunWithSchema(schema Schema, db *sqlx.DB, run func(db *sqlx.DB)) {
	runner := func(create, drop string) {
		defer func() { MultiExec(db, drop) }()
		MultiExec(db, create)
		run(db)
	}

	if Postgres {
		create, drop := schema.Postgres()
		runner(create, drop)
	}
	if Sqlite {
		create, drop := schema.Sqlite3()
		runner(create, drop)
	}
	if Mysql {
		create, drop := schema.MySQL()
		runner(create, drop)
	}
}
