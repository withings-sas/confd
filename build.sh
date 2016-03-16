#!/usr/bin/env bash

export GOPATH=$1

# golang installed from ppa:ubuntu-lxc/lxd-stable

go get github.com/BurntSushi/toml
go get github.com/go-sql-driver/mysql
go get github.com/kelseyhightower/confd

mkdir -p mkdir go/src/git.corp.withings.com
ln -s $(pwd) go/src/git.corp.withings.com/confd

echo "Building confd..."
[[ -d bin ]] && rm -rf bin
mkdir -p bin
go build -o bin/confd .
