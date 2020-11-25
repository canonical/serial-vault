[![Build Status][travis-image]][travis-url]
# Serial Vault

A Go web service that digitally signs device assertion details.

The application can be run in three modes: signing service, admin service and system-user assertion services. All the web services
operate under unencrypted HTTP connections, so these should not be exposed to a public network
as-is. The services should be protected by web server front-end services, such as Apache, that
provide secure HTTPS connections. Also, the admin service does not include authentication nor 
authorisation, so this service will typically be made available on a restricted network with some 
authentication front-end on the web server e.g. SSO. Typically, the services will only be available 
on a restricted network at a factory, though, with additional security measures, the signing service 
could be made available on a public network.

Some deployment recommendations are [provided](docs/installation.md)

The service mode (signing or admin) is defined in the settings.yaml file. The
selected service should be accessible on port :8080 or :8081:
 - Signing/API Service: http://localhost:8080/v1/version
 - Admin/UI Service: http://localhost:8081/

The Admin service's CSRF protection sends a cookie over a secure channel. If the cookie is to be sent
over an insecure channel, it is needed to workaround it by setting the environment variable:
```bash
$ export CSRF_SECURE=disable
```
When modified that environment variable is set, consider that the current web session must be invalidated
in order to changes take effect. That could require a browser restart.
NEVER set this configuration in production environments.

## Install from Source
If you have a Go development environment set up, we recommend at least Go v1.13 or higher. This project is built
and tested on `Ubuntu 16.04.7 LTS (Xenial Xerus)` with `go 1.13`. If you wish to build or run the service from source we
recommend using LXD container for this purpose. To get started with LXD, please follow this 
[wiki](https://wiki.canonical.com/UbuntuOne/Developer/LXC) (jump to the LXD section):

  ```bash
  lxc launch ubuntu:16.04 serial-vault -p default -p $USER
  lxc ls
  ssh -A <container-ip>
  serial-vault:~$ git clone https://github.com/CanonicalLtd/serial-vault.git
  serial-vault:~$ cd serial-vault
  serial-vault:~/serial-vault$ sudo ./setup-container
  serial-vault:~/serial-vault$ make bootstrap
  serial-vault:~/serial-vault$ make install
  ```

After this you will find all binaries in the `bin/` folder of the project.

```bash
@serial-vault:~/serial-vault$ tree bin/
bin/
├── factory
├── serial-vault
├── serial-vault-admin
```

To run Serial Vault in admin/UI mode you will need a `static/` folder. Make sure to setup 
correct path to this folder in the configuration file with `docRoot` variable.

### Configuration
- Install PostgreSQL and create a database.
- Set up the config file, using `settings.yaml.example` as a guide.
- Create the database tables:

  ```bash
  $ cd serial-vault
  $ go get ./...
  $ make migrate
  ```

Sample Serial Vault Configuration:
```
title: "Serial Vault"
logo: "/static/images/logo-ubuntu-white.svg"

# Path to the assets (${docRoot}/static)
docRoot: "."

# Backend database details
driver: "postgres"
datasource: "postgres://vault:vault@localhost:5432/vault?sslmode=disable"

keystore: "database"
keystoreSecret: "KEYSTORE_SECRET"

# Valid API keys
apiKeys:
  - API_KEY

# 32 bytes long key to protect server from cross site request forgery attacks
csrfAuthKey: "32_BYTES_LONG_CSRF_AUTH_KEY"
```

### Run the service
  ```bash
  $ cd $GOPATH/src/github.com/CanonicalLtd/serial-vault
  # run the service in sign/API mode
  $ make run-sign
  # run the serivce in admin/UI
  $ make run-admin
  ```

## Deploy it with Juju
Juju greatly simplifies the deployment of the Serial Vault. A charm bundle is available
at the [charm store](https://jujucharms.com/u/canonical-solutions/serial-vault-bundle/), which deploys
everything apart from the Apache front-end units. There is an example of using Juju in the
[Deployment Guidelines](docs/installation.md).

## Try with docker

  ```bash
  $ git clone https://github.com/CanonicalLtd/serial-vault
  $ make run-docker
  # remove containers after try
  $ make stop-docker
  ```

## Development Environment

### Contributing

The general workflow is [forking](https://help.github.com/en/github/getting-started-with-github/fork-a-repo) the Serial Vault GitHub repository, 
make changes in a branch and then create a [pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork).

#### Adding new golang dependency

`Serial Vault` uses `go mod` to manage its dependencies.

- Run `go get foo/bar` in the source folder to add the dependency to go.mod file.
- Run `go build ./...` to check that everything works.

To remove a dependency

- Edit your code and remove the import reference.
- Run `go mod tidy` in the source folder to remove dependency from go.mod file.

### Install Go
Follow the instructions to [install Go](https://golang.org/doc/install). The current version of the service runs with `Go 1.13`.

### Install the React development environment
#### Pre-requisites
- Install the build packages
```bash
sudo apt-get install build-essential libssl-dev
# For TPM2.0 (optional)
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
nvm install lts/*

# Select the version to use
nvm ls
nvm use lts/*
```

- Install the nodejs dependencies
```bash
cd serial-vault
npm install
```

### Working with React

The frontend code can be found in `webapp-admin` directory.

#### Building static assets locally

```bash
# Select the version to use
cd webapp-admin/
nvm ls
nvm use lts/*
npm run build
```

#### Production static assets build process

Production build for the frontend part (javascript and css) is semi-automated and done with [GitHub Actions](https://github.com/features/actions). 
You can find the configuration for this process in `.github/workflows/nodejs.yml`. The build process starts automatically after the PR is approved 
and pushed to `master`. 

You can see the build process in [actions](https://github.com/CanonicalLtd/serial-vault/actions) tab of this project.
After the successful build the automation bot will create a PR with the generated build artifact (minified javascript code) in the `static/` 
directory of this project. These PRs can be merged manually.

#### Run the tests
```bash
npm test
```


## API Methods

### /v1/version (GET)
> Return the version of the serial vault service.

#### Output message
```json
{
  "version":"2.1-0",
}
```
- version: the version of the serial vault service (string)


### /v1/request-id (POST)
> Returns a nonce that is needed for the 'serial' request.

#### Output message
```json
{
  "request-id": "abc123456",
  "success": true,
  "message": ""
}
```
- success: whether the request was successful (bool)
- message: error message from the request (string)
- request-id: unique string that is needed for serial requests (string)

The request-id is a nonce that can only be used once and must be used before it expires (typically 600 seconds).

### /v1/serial (POST)
> Generate a serial assertion signed by the brand key.

Takes the details from the device as a serial-request assertion and generates a signed serial assertion.

#### Input message
The message must be the serial-request assertion format and is best generated using the snapd libraries.
```
type: serial-request
brand-id: System
model: Router 3400
device-key:
    WkUDQbqFCKZBPvKbwR...
request-id: abc123456
body-length: 10
sign-key-sha3-384: UytTqTvREVhx...

HW-DETAILS
serial: A1228ML

AcLBUgQAAQoABgUCV7R2C...
```
- brand-id: the Account ID of the manufacturer (string)
- model: the name of the device (string)
- device-key: the encoded type and public key of the device (string)
- request-id: the nonce returned from the /v1/nonce method (string)
- signature: the signed data
- serial: serial number of the device (string)

The HW-DETAILS are optional hardware details in YAML format, but must include the 'serial' tag as that is a mandatory
part of the serial assertion.

#### Output message
The method returns a signed serial assertion using the key from the vault.

### /v1/pivot (POST)
> Find the model pivot details for a device.

Takes the serial assertion from a device and returns the model pivot details for the device.

#### Input message
The serial assertion of the manufacturer's device.

#### Output message
The method returns details of the model pivot, to convert the device to a reseller model.

### /v1/pivotmodel (POST)
> Generate the model assertion for the pivoted model for a device.

Takes the serial assertion from a device and returns the model assertion for the pivoted model of the device.

#### Input message
The serial assertion of the manufacturer's device.

#### Output message
The method returns details of the model assertion of the pivoted model, to convert the device to a reseller model.

### /v1/pivotserial (POST)
> Generate the serial assertion for the pivoted model for a device.

Takes the serial assertion from a device and returns the serial assertion for the pivoted model of the device.

#### Input message
The serial assertion of the manufacturer's device.

#### Output message
The method returns details of the serial assertion of the pivoted model, to convert the device to a reseller model.

[travis-image]: https://travis-ci.org/CanonicalLtd/serial-vault.svg?branch=master
[travis-url]: https://travis-ci.org/CanonicalLtd/serial-vault
[actions]: https://github.com/CanonicalLtd/serial-vault/actions
