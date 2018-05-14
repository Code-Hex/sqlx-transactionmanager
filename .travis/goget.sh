#!/bin/bash
set -ev

go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls
go get github.com/go-sql-driver/mysql
go get github.com/jmoiron/sqlx
github.com/mattn/go-sqlite3