[![Build Status][travis-image]][travis-url]
# Identity Vault

A Go web service that digitally signs device assertion details.

## Install
Go get it:

  ```bash
  $ go get github.com/ubuntu-core/identity-vault
  ```

Configure it:
- Install PostgreSQL and create a database.
- Set up the config file, using ```settings.yaml``` as a guide.
- Create the database tables:
  ```bash
  $ cd identity-vault
  $ go run tools/createdb.go
  ```

Run it:
  ```bash
  $ cd identity-vault
  $ go run server.go -config=/path/to/settings.yaml -mode=signing
  ```

The application has an admin service that can be run by using mode=admin.

## Development Environment

### Install Go
Follow the instructions to [install Go](https://golang.org/doc/install).

### Install the React development environment
#### Pre-requisites
- Install the build packages
```bash
sudo apt-get install build-essential libssl-dev
# For TPM2.0
sudo apt-get install tpm2-tools
```

- Install NVM
Install the [Node Version Manager](https://github.com/creationix/nvm) that will allow a specific
version of Node.js to be installed. Follow the installation instructions.

- Install the latest stable Node.js and npm
The latest stable (LTS) version of Node can be found on the [Node website](nodejs.org).
```bash
# Overview of available commands
nvm help

# Install the latest stable version
nvm install v4.4.3

# Select the version to use
nvm ls
nvm use v4.4.3

# Install gulp globally
npm install -g gulp
```

- Install the nodejs dependencies
```bash
cd identity-vault
npm install
```

### Working with React

#### Build the project bundle
```bash
# Select the version to use
nvm ls
nvm use v4.4.3
gulp
```

#### Run the tests
```bash
npm test
```


## API Methods

### /1.0/version (GET)
> Return the version of the identity vault service.

#### Output message
```json
{
  "version":"0.1.0",
}
```
- version: the version of the identity vault service (string)


### /1.0/models (GET)
> Return the available models from the identity vault.

#### Output message
```json
{
  "success": true,
  "message": "",
  "models": [
  {
    "brand-id": "System",
    "model": "DroidBox 2400",
    "type": "device",
    "revision": 2
  },
  {
    "brand-id": "System",
    "model": "DroidBox 1200",
    "type": "device",
    "revision": 1
  },
  {
    "brand-id": "System",
    "model": "Drone 1000",
    "type": "device",
    "revision": 4
  }]
}
```
- success: whether the request was successful (bool)
- message: error message from the request (string)
- models: the list of available models (array)


### /1.0/sign (POST)
> Clear-sign the device identity details.

Takes the details from the device, formats the data and clear-signs it.

#### Input message
The message must be the serial assertion format and is best generated using the snapd libraries.
```
type: serial
authority-id: System
brand-id: System Inc.
model: Router 3400
revision: 12
serial: A1228M\L
timestamp: 2016-01-02T15:04:05Z
device-key: openpgp WkUDQbqFCKZBPvKbwR...

openpgp mQINBFaiIK4BEADHpUm...
```
- brand-id: the Account ID of the manufacturer (string)
- model: the name of the device (string)
- serial: serial number of the device (string)
- device-key: the type and public key of the device (string)
- revision: the revision of the device (integer)
- signature: the signed data

#### Output message
The method returns a signed serial assertion using the key from the vault.


[travis-image]: https://travis-ci.org/ubuntu-core/identity-vault.svg?branch=master
[travis-url]: https://travis-ci.org/ubuntu-core/identity-vault
