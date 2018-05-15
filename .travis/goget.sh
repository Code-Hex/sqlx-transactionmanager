#!/bin/bash
set -ev

go get -v golang.org/x/tools/cmd/cover
go get -v github.com/mattn/goveralls
go get -v github.com/lib/pq
go get -v github.com/go-sql-driver/mysql
go get -v github.com/jmoiron/sqlx
go get -v github.com/mattn/go-sqlite3