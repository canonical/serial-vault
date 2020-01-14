#!/bin/sh -u

set +e

# Polling PostgreSQL server
while true ; do
  result=$(psql -lqt postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT | cut -d \| -f 1 | grep -qw $POSTGRES_DB)
  ret=$?
  if [ $ret -ne 0 ]; then
    echo "Waiting database creation on PostgreSQL server."
    sleep 3
  else
    break
  fi
done

set -e

cd $GOPATH/src/github.com/CanonicalLtd/serial-vault

sed -i  \
  -e "s/API_KEY/$API_KEY/g" \
  -e "s/POSTGRES_HOST/$POSTGRES_HOST/g" \
  -e "s/POSTGRES_DB/$POSTGRES_DB/g" \
  -e "s/POSTGRES_PASSWORD/$POSTGRES_PASSWORD/g" \
  -e "s/POSTGRES_PORT/$POSTGRES_PORT/g" \
  -e "s/POSTGRES_USER/$POSTGRES_USER/g" \
  -e "s/KEYSTORE_SECRET/$KEYSTORE_SECRET/g" \
  settings.yaml

echo "Apply database migrations"
go run cmd/serial-vault-admin/main.go database

echo "Starting server"
go run cmd/serial-vault/main.go -mode=admin &
go run cmd/serial-vault/main.go -mode=signing
