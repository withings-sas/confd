#!/usr/bin/env bash

export GOPATH=$1

go get github.com/BurntSushi/toml
go get github.com/go-sql-driver/mysql

echo "Building confd..."
mkdir -p bin
go build -o bin/confd .
