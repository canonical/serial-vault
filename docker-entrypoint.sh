#!/bin/sh
set +x
set -e

cd $GOPATH/src/github.com/ubuntu-core/identity-vault
go run tools/createdb.go

go run server.go -config=settings.yaml
