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
# sudo is only needed if you are not logged into the store
sudo snap install --dangerous serial-vault-services_*_amd64.snap
```

## Configure the Serial Vault Services
To initialize the database and to show the settings.yaml configuration file:

```bash
cd serial-vault-services-snap  # Make sure you are in the correct directory
serial-vault-services.serviceinit
```

This will display the settings.yaml file for the serial vault. To use the details, copy and
paste them into a settings.yaml file and then run:
```bash
cat settings.yaml | sudo serial-vault-services.config
sudo snap disable serial-vault-services
sudo snap enable serial-vault-services
```

The services are then accessible via:
Admin Service   : http://localhost/admin/
User Service    : http://localhost/user/
Signing Service : http://localhost/signing/v1/version


## Set-up Apache and SSL
Apache is configured to use a self-signed certificate, which will cause a browser warning
for the certificate. It is possible to supply your own certificate using the ```enable-https``` command.

```bash
sudo serial-vault-services.enable-https -h
```
