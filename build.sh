#!/usr/bin/env bash

export GOPATH=$1

# golang installed from ppa:ubuntu-lxc/lxd-stable

go get github.com/BurntSushi/toml
go get github.com/go-sql-driver/mysql
go get github.com/kelseyhightower/confd
ln -s . go/src/git.corp.withings.com/confd

echo "Building confd..."
mkdir -p bin
go build -o bin/confd .
