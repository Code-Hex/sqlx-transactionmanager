# sqlx-transactionmanager

Transaction handling for database it extends https://github.com/jmoiron/sqlx

[![GoDoc](https://godoc.org/github.com/Code-Hex/sqlx-transactionmanager?status.svg)](https://godoc.org/github.com/Code-Hex/sqlx-transactionmanager) 
[![Build Status](https://travis-ci.org/Code-Hex/sqlx-transactionmanager.svg?branch=master)](https://travis-ci.org/Code-Hex/sqlx-transactionmanager) 
[![Coverage Status](https://coveralls.io/repos/github/Code-Hex/sqlx-transactionmanager/badge.svg?branch=master)](https://coveralls.io/github/Code-Hex/sqlx-transactionmanager?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/sqlx-transactionmanager)](https://goreportcard.com/report/github.com/Code-Hex/sqlx-transactionmanager)

## Synopsis

```go
db := sqlx.MustOpen("mysql", dsn())

// starts transaction statements
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

tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "Code", "Hex", "x00.x7f@gmail.com")
tx.MustExec("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?", "a@b.com", "Code", "Hex")

var p Person
if err := tx.Get(&p, "SELECT * FROM person LIMIT 1"); err != nil {
    panic(err)
}

// transaction commits
if err := tx.Commit(); err != nil {
    panic(err)
}

fmt.Println(p)
```

## Description

sqlx-transactionmanager is a simple transaction manager. This package provides nested transaction management on multi threads.

See more details [example code](https://github.com/Code-Hex/sqlx-transactionmanager/blob/master/eg/main.go#L57-L87) if you want to know how to use this.

## Install

    go get github.com/Code-Hex/sqlx-transactionmanager

## Contributing

I'm looking forward you to send pull requests or reporting issues.

## License

[MIT](https://github.com/Code-Hex/sqlx-transactionmanager/blob/master/LICENSE)

## Author

[CodeHex](https://twitter.com/CodeHex)  