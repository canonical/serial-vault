---
title: Serial Vault Snap
table_of_contents: True
---

# Overview

Are you using one of the many systems that support snaps?. 
A pre-built snap is available from the Snap store.
You can easily install it by typing 

```
sudo snap install serial-vault
```

The service needs a PostgreSQL database instance to run so, you'll also
need to:

 * Install PostgreSQL and create a database
 * Set up the config file, using _settings.yaml_ in this project root folder as guide
 * Configure the snap using the _settings.yaml_ file as input

 The snap will create or update the tables in the database on restart, as soon as it has
 a valid database connection.

 Depending on the service mode (signing, admin or system-user) defined in the settings.yaml
 file, the selected mode should be accessible on port :8080, :8081 or :8082

| Service | URL |
|---------|-----|
| Signing Service | http://localhost:8080/v1/version |
| Admin Service | http://localhost:8081/ |
| System-User Service | http://localhost:8082 |


NOTE: In the specific case of the Admin Service, due to cross site request forgery protection
there is a need of accessing through a https frontend. A simple Apache reverse proxy
can be used for this. The only requirement is to securize the access using a certificate
whose issued common name matches the requested DNS hostname.
For example, having a certificate having

```
cn=the_host
```

it is needed that

```
https://the_host/
```

resolves to the apache frontend

## Example

### Install PostgreSQL and create database

Next are explained the steps to install and configure a database in PostgreSQL. This 
is not the only or best recommended way to do it. Just another one.
Install it in Ubuntu classic by typing:

```
sudo apt update
sudo apt install postgresql postgresql-contrib
```

Create a user to access the database:
```
sudo -i -u postgres
createuser --interactive
Enter name of the role to add: THEUSER
Shall the new role be a superuser? (y/n): Y
```

Now, create the database:
```
psql -d serialvault
alter user THEUSER with password 'THEPASSWORD';
```

### Install and configure the snap

```
sudo snap install serial-vault
```

Create a snapcraft.yaml like:

```
title: "Serial Vault"
logo: "/static/images/logo-ubuntu-white.svg"

# Service mode: signing or admin or system-user
mode: admin

# Path to the assets (${docRoot}/static)
docRoot: "."

# Backend database details
driver: "postgres"
datasource: "dbname=serialvault sslmode=disable user=THEUSER password=THEPASSWORD"

# For Database
keystore: "database"
keystoreSecret: "secret code to encrypt the auth-key hash"

# 32 bytes long key to protect server from cross site request forgery attacks
csrfAuthKey: "2E6ZYnVYUfDLRLV/ne8M6v1jyB/376BL9ORnN3Kgb04uSFalr2ygReVsOt0PaGEIRuID10TePBje5xdjIOEjQQ=="

# Return URL of the service (needed for OpenID)
urlHost: "serial-vault"
urlScheme: http

# Enable user authentication using Ubuntu SSO
enableUserAuth: True
```

inject the configuration by using config app of the snap, and restart service to apply:

```
cat /path/to/settings.yaml | sudo /snap/bin/serial-vault.config
sudo systemctl restart snap.serial-vault.service.service
```

Deploy and configure an apache https frontend at _serial-vault_ dns et voilá!. You should see 
the admin service, in this case, by accessing:

```
https://serial-vault
```
