# Factory Serial Vault

The full suite of Serial Vault Services in the factory:

 - Signing Service: web service to sign the serial assertion for a device
 - Sync Service: scheduled task that syncs with a cloud Serial Vault

## Build the Factory Serial Vault Snap
Install [snapcraft](https://snapcraft.io/)

```bash
snapcraft cleanbuild
```

## Install the Factory Serial Vault
```bash
# sudo is only needed if you are not logged into the store
sudo snap install --dangerous factory-serial-vault_*_amd64.snap
```

## Configure the Factory Serial Vault
To generate a `settings.yaml` configuration file:

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

The local sqlite3 database will be generated and synchronized with the cloud serial vault. 
The factory database will include all the data needed to provide signed serial assertions 
to devices in the factory.

The services are then accessible via:
Signing Service : http://localhost/v1/version


## Set-up Apache and SSL
Apache is configured to use HTTP by default. It is possible to use HTTPS by generating a self-signed certificate or
supplying your own certificate using the ```enable-https``` command.
```bash
sudo factory-serial-vault.enable-https -h
```
Note: snapd will not be able to use the signing API if a self-signed certificate is used.
