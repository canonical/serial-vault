# Factory Serial Vault

The full suite of Serial Vault Services:

 - Signing Service: web service to sign the serial assertion for a device

## Build the Serial Vault Services Snap
Install [snapcraft](https://snapcraft.io/)

```bash
cd factory-serial-vault  # Make sure you are in the correct directory
snapcraft
```

## Install the Serial Vault Services
The Factory Serial Vault needs to have a PostgreSQL service installed and configured. The 
installation script simplifies this process by handling the full installation process.

```bash
cd factory-serial-vault  # Make sure you are in the correct directory
# sudo is only needed if you are not logged into the store
sudo snap install --dangerous factory-serial-vault_*_amd64.snap
```

## Configure the Factory Serial Vault
To initialize the database and to show the settings.yaml configuration file:

```bash
factory-serial-vault.serviceinit
```

This will display the settings.yaml file for the serial vault. To use the details, copy and
paste them into a settings.yaml file and then run:
```bash
cat settings.yaml | sudo factory-serial-vault.config
sudo snap disable factory-serial-vault
sudo snap enable factory-serial-vault
```

Note: 
 - The database will be initialized and run as the local user, so it is recommended that a user is created specifically to run the database service.
 - The ```serviceinit``` command generates a keystore secret and that is used to encrypt the private keys that are stored in the database. It is important, to backup the keystore secret as well as the database so that the services can be recovered successfully.

The services are then accessible via:
Signing Service : http://localhost/v1/version


## Set-up Apache and SSL
Apache is configured to use HTTP by default. It is possible to use HTTPS by generating a self-signed certificate or
supplying your own certificate using the ```enable-https``` command.
```bash
sudo factory-serial-vault.enable-https -h
```
Note: snapd will not be able to use the signing API if a self-signed certificate is used.

## Restarting Services
The factory serial vault will run automatically when the system reboots, apart from the database. The database
runs as the local user, so it cannot be started as a snappy daemon (which run as root). Consequently, the
database will need to be restarted manually:

```bash
factory-serial-vault.startdb
sudo snap disable factory-serial-vault
sudo snap enable factory-serial-vault

factory-serial-vault.statusdb
```