#!/bin/sh

# Uninstall the services
snap remove serial-vault-services

# Uninstall and remove the database
dropdb serialvault
dropuser serialvault
snap remove postgresql96

# Remove the user
userdel serialvault
