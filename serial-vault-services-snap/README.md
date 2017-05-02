# Serial Vault Services

The full suite of Serial Vault Services:

 - Admin Service: manage models and signing keys
 - Signing Service: web service to sign the serial assertion for a device
 - System-User Service: web service to create a system-user assertion for a device

## Build the Serial Vault Services Snap
Install [snapcraft](https://snapcraft.io/)

```bash
cd serial-vault-services-snap  # Make sure you are in the correct directory
snapcraft
```

## Install the Serial Vault Services
The Serial Vault Services need to have a PostgreSQL service installed and configured. The 
installation script simplifies this process by handling the full installation process.

```bash
cd serial-vault-services-snap  # Make sure you are in the correct directory
sudo ./install.sh
```

## Uninstall the Serial Vault Services
A script has been provided to uninstall the services and remove the databases and the
user that were created. The database is not backed up, so make sure that you do not
have data that needs to be kept within the database.

```bash
cd serial-vault-services-snap  # Make sure you are in the correct directory
sudo ./uninstall.sh
```
