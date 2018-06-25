# sqlx-transactionmanager

Transaction handling for database it extends https://github.com/jmoiron/sqlx

[![GoDoc](https://godoc.org/github.com/Code-Hex/sqlx-transactionmanager?status.svg)](https://godoc.org/github.com/Code-Hex/sqlx-transactionmanager) 
[![Build Status](https://travis-ci.org/Code-Hex/sqlx-transactionmanager.svg?branch=master)](https://travis-ci.org/Code-Hex/sqlx-transactionmanager) 
[![Coverage Status](https://coveralls.io/repos/github/Code-Hex/sqlx-transactionmanager/badge.svg?branch=master)](https://coveralls.io/github/Code-Hex/sqlx-transactionmanager?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/sqlx-transactionmanager)](https://goreportcard.com/report/github.com/Code-Hex/sqlx-transactionmanager)

## Synopsis

<details>
  <summary>Standard</summary>

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
    return err
}

// transaction commits
if err := tx.Commit(); err != nil {
    return err
}

fmt.Println(p)
```
</details>

<details>
  <summary>Nested Transaction</summary>

```go
db := sqlx.MustOpen("mysql", dsn())

func() {
    // We should prepare to recover from panic.
    defer func() {
        if r := recover(); r != nil {
            // Do something recover process
        }
    }()
    // Start nested transaction.
    // To be simple, we will cause panic if something sql process if failed.
    func() {
        // starts transaction statements
        tx, err := db.BeginTxm()
        if err != nil {
            panic(err)
        }
        // Do rollbacks if fail something in nested transaction.
        defer tx.Rollback()
        func() {
            // You don't need error handle in already began transaction.
            tx2, _ := db.BeginTxm()
            defer tx2.Rollback()
            tx2.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "Code", "Hex", "x00.x7f@gmail.com")
            // Do something processing.
            // You should cause panic() if something failed.
            if err := tx2.Commit(); err != nil {
                panic(err)
            }
        }()
        tx.MustExec("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?", "a@b.com", "Code", "Hex")
        if err := tx.Commit(); err != nil {
            panic(err)
        }
    }()
}()

var p Person
if err := tx.Get(&p, "SELECT * FROM person LIMIT 1"); err != nil {
    return err
}

fmt.Println(p)
```
</details>

<details>
  <summary>Transaction block</summary>

```go
var p Person
if err := tm.Run(db, func(tx tm.Executor) error {
    _, err := tx.Exec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "Al", "Paca", "x00.x7f@gmail.com")
    if err != nil {
        return err
    }
    _, err = tx.Exec("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?", "x@h.com", "Al", "Paca")
    if err != nil {
        return err
    }

    return tx.QueryRow("SELECT * FROM person LIMIT 1").Scan(&p.FirstName, &p.LastName, &p.Email, &p.AddedAt)
}); err != nil {
    panic(err)
}
println(&p)

if err := tm.Runx(db, func(tx tm.Executorx) error {
    tx.MustExec(tx.Rebind("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)"), "Code", "Hex", "x00.x7f@gmail.com")
    tx.MustExec(tx.Rebind("UPDATE person SET email = ? WHERE first_name = ? AND last_name = ?"), "a@b.com", "Code", "Hex")
    if err := tx.Get(&p, "SELECT * FROM person ORDER BY first_name DESC LIMIT 1"); err != nil {
        return err
    }
    return nil
}); err != nil {
    panic(err)
}
println(&p)
```
</details>

## Description

sqlx-transactionmanager is a simple transaction manager. This package provides nested transaction management on multi threads.

See more details [example for extends sqlx](https://github.com/Code-Hex/sqlx-transactionmanager/blob/master/eg/main.go#L57-L87) or [example for transaction block](https://github.com/Code-Hex/sqlx-transactionmanager/blob/master/eg/tm/main.go#L58-L90) if you want to know how to use this.

## Install

    go get github.com/Code-Hex/sqlx-transactionmanager

## Contributing

I'm looking forward you to send pull requests or reporting issues.

## License

[MIT](https://github.com/Code-Hex/sqlx-transactionmanager/blob/master/LICENSE)

## Author

[CodeHex](https://twitter.com/CodeHex)  