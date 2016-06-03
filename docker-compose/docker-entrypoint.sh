#!/bin/sh
set +e

# polling postgresql server
while true ; do
	result=$(psql -lqt postgres://postgres:ubuntu@db:5432 | cut -d \| -f 1 | grep -qw identityvault)
	ret=$?
	if [ $ret -ne 0 ]; then
		echo "waiting database creation on postgresql server"
		sleep 3
	else
		break
	fi
done

set -e

cd $GOPATH/src/github.com/ubuntu-core/identity-vault
go run tools/createdb.go -config=docker-compose/settings.yaml

go run server.go -config=docker-compose/settings.yaml
