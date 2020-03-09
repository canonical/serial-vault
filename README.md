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
If you have a Go development environment set up, Go get it, we recommend at least Go v1.13 or higher.

  ```bash
  $ go get github.com/CanonicalLtd/serial-vault/...
  ```

### Configure it:
- Install PostgreSQL and create a database.
- Set up the config file, using ```settings.yaml``` as a guide.
- Create the database tables:
  ```bash
  $ cd serial-vault
  $ ./get-deps.sh
  $ go run cmd/serial-vault-admin/main.go database --config=/path/to/settings.yaml
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

### Run it:
  ```bash
  $ cd $GOPATH/src/github.com/CanonicalLtd/serial-vault
  $ go run cmd/serial-vault/main.go -config=/path/to/settings.yaml -mode=signing
  ```

The application has an admin/UI service that can be run by using mode=admin.

## Deploy it with Juju
Juju greatly simplifies the deployment of the Serial Vault. A charm bundle is available
at the [charm store](https://jujucharms.com/u/canonical-solutions/serial-vault-bundle/), which deploys
everything apart from the Apache front-end units. There is an example of using Juju in the
[Deployment Guidelines](docs/installation.md).

## Try with docker

  ```bash
  $ git clone https://github.com/CanonicalLtd/serial-vault
  $ cd serial-vault/docker-compose
  $ docker-compose up
  # remove containers after try
  $ docker-compose kill && docker-compose rm
  ```

## Development Environment

### Contributing

The general workflow is forking the Serial Vault GitHub repository, make changes in a branch and then create a pull request:

- Pull the original package:
  `go get github.com/CanonicalLtd/serial-vault/...`
- [Fork](https://github.com/CanonicalLtd/serial-vault/fork) the Serial Vault repository on GitHub
- Change to the top level of the repository
  `cd $GOPATH/src/github.com/CanonicalLtd/serial-vault`
- Add your fork
  `git remote add fork https://github.com/username/serial-vault`
- Create a feature branch
  `git checkout -b new-feature`
- Commit your changes to your forked repo
  `git commit -am "New Feature"`
  and
  `git push fork`
- Follow the link from the cli to create new PR on GitHub.

#### Adding new golang dependency

We are using `govendor` tool to manage dependency in Serial Vault. It will be installed after the first run of `get-deps.sh`. If you need to add a new dependency to this project, please run  `govendor fetch github.com/new/package` and commit the changes in `vendor/vendor.json` file.

### Install Go
Follow the instructions to [install Go](https://golang.org/doc/install).

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
You can find the configuration for this process in `.github/workflows/nodejs.yml`. The build process starts automatically after the PR is approved and pushed to `master`. 
You can see the build process in [actions](https://github.com/CanonicalLtd/serial-vault/actions) tab of this project. 
After the successful build the automation bot will create a PR with the generated build artifact (minified javascript code) in the `static/` directory of this project. These PRs can be merged manually.

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
