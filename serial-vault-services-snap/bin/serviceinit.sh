#!/bin/sh

# The PostgreSQL snap only works with the en_US locale
LC_ALL=en_US.UTF-8

# Initialize and start the database service
echo "Initialize the database"
$SNAP/usr/bin/wrapper-initialize initdb
$SNAP/usr/bin/wrapper-pg_ctl -D $SNAP_USER_COMMON/data -l $SNAP_USER_COMMON/logs/logfile start

# Wait until the database is ready
while ! $SNAP/usr/bin/pg_isready -h 127.0.0.1; do
    sleep 1
done

# # Generate the password for the database user
# db_password=$(cat /dev/urandom | tr -dc _A-Z-a-z-0-9 | head -c64)

# Create the user and the database
$SNAP/usr/bin/createdb -h 127.0.0.1 $USER

# # Update the database user's password
# $SNAP/usr/bin/wrapper-psql -h 127.0.0.1 $USER <<SQL
# ALTER USER $USER WITH PASSWORD '$db_password';
# SQL

# Generate the serial-vault secrets and API key
keystore_secret=$(cat /dev/urandom | tr -dc A-Z-a-z-0-9 | head -c64)
api_key=$(cat /dev/urandom | tr -dc A-Z-a-z-0-9 | head -c64)
csrf_key=$(cat /dev/urandom | tr -dc A-Z-a-z-0-9 | head -c64)


# Generate the configuration for the serial-vault
echo "\n\nSave the config file settings to a file e.g. settings.yaml"
echo "================================================"
cat <<SETTINGS
docRoot: "."

driver: "postgres"
datasource: "dbname=$USER user=$USER sslmode=disable"

keystore: "database"
keystoreSecret: "$keystore_secret"

apiKeys:
    - $api_key

csrfAuthKey: "$csrf_key"
SETTINGS
echo "================================================"

echo "\nConfigure the serial vault services:"
echo "cat settings.yaml | sudo serial-vault-services.config"

echo "\nRestart the services to complete the initialization"
echo "sudo snap disable serial-vault-services"
echo "sudo snap enable serial-vault-services"
echo "\n-------------------------------------------------"
echo "Admin Service   : http://localhost:8081"
echo "User Service    : http://localhost:8082"
echo "Signing Service : http://localhost:8080/v1/version"
