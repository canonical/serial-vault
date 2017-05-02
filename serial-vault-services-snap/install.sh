#!/bin/sh

adduser --disabled-login --disabled-password --gecos "" serialvault

# Install postgresql
snap install postgresql96

# Generate the password for the database user
db_password=$(cat /dev/urandom | tr -dc _A-Z-a-z-0-9 | head -c64)

# Initialise and start the database service
su - serialvault -c "postgresql96.initialize initdb"
su - serialvault -c "postgresql96.pgctl -D /home/serialvault/snap/postgresql96/common/data -l /home/serialvault/snap/postgresql96/common/logs/logfile start"


# Create the user and database
postgresql96.createuser -s -d -h 127.0.0.1 serialvault
su - serialvault -c "postgresql96.createdb -h 127.0.0.1 serialvault"
postgresql96.psql -h 127.0.0.1 serialvault <<SQL
ALTER USER serialvault WITH PASSWORD '$db_password';
SQL

# Install the serial-vault services snap
snap install --dangerous serial-vault-services_1.5_amd64.snap


# Generate the serial-vault secrets and API key
keystore_secret=$(cat /dev/urandom | tr -dc _A-Z-a-z-0-9 | head -c64)
api_key=$(cat /dev/urandom | tr -dc A-Z-a-z-0-9 | head -c64)
csrf_key=$(cat /dev/urandom | tr -dc _A-Z-a-z-0-9 | head -c64)

echo "\nCredentials:"
echo "Database Password: $db_password"
echo "Keystore Secret: $keystore_secret"
echo "API Key: $api_key"
echo "CSRF Auth Key: $csrf_key"

# Configure the serial-vault
/snap/bin/serial-vault-services.config <<SETTINGS
docRoot: "."

driver: "postgres"
datasource: "dbname=serialvault user=serialvault password=$db_password"

keystore: "database"
keystoreSecret: "$keystore_secret"

# Valid API keys
apiKeys:
    - $api_key

csrfAuthKey: "$csrf_key"
SETTINGS

systemctl restart snap.serial-vault-services.admin.service
systemctl restart snap.serial-vault-services.signing.service
systemctl restart snap.serial-vault-services.user.service

echo "\n-------------------------------------------------"
echo "Admin Service   : http://localhost:8081"
echo "User Service    : http://localhost:8082"
echo "Signing Service : http://localhost:8080/v1/version"
