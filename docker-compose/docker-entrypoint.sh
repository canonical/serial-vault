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

cd /go/src/github.com/CanonicalLtd/serial-vault

sed -i  \
  -e "s/API_KEY/$API_KEY/g" \
  -e "s/POSTGRES_HOST/$POSTGRES_HOST/g" \
  -e "s/POSTGRES_DB/$POSTGRES_DB/g" \
  -e "s/POSTGRES_PASSWORD/$POSTGRES_PASSWORD/g" \
  -e "s/POSTGRES_PORT/$POSTGRES_PORT/g" \
  -e "s/POSTGRES_USER/$POSTGRES_USER/g" \
  -e "s/KEYSTORE_SECRET/$KEYSTORE_SECRET/g" \
  settings.yaml

/usr/lib/go-1.13/bin/go run cmd/serial-vault/main.go -mode=admin &
/usr/lib/go-1.13/bin/go run cmd/serial-vault/main.go -mode=signing
