#!/bin/sh

# Generate the serial-vault secrets
keystore_secret=$(cat /dev/urandom | tr -dc A-Z-a-z-0-9 | head -c64)

# Generate the configuration for the serial-vault
echo "\n\nSave the config file settings to a file e.g. settings.yaml"
echo "================================================"
cat <<SETTINGS
docRoot: "."

driver: "sqlite3"
datasource: "/var/snap/factory-serial-vault/current/factory.db"

keystore: "database"
keystoreSecret: "$keystore_secret"

# Factory sync - CHANGEME
syncUrl: "https://serial-vault-partners.canonical.com/api/"
syncUser: "lpuser"
syncAPIKey: "user-apikey"
SETTINGS
echo "================================================"

echo "\nConfigure the serial vault services:"
echo "cat settings.yaml | sudo factory-serial-vault.config"

echo "\nRestart the services to complete the initialization"
echo "sudo snap disable factory-serial-vault"
echo "sudo snap enable factory-serial-vault"
echo "\n-------------------------------------------------"
echo "Signing Service : http://localhost/v1/version"