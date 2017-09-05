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

Some deployment recommendations are [provided](docs/Deployment.md)

## Install using the snap package
Are you using one of the many systems that support [snaps](https://snapcraft.io/)?
A pre-built snap is available from the [Snap store](https://uappexplorer.com/app/serial-vault.james).

```bash
$ sudo snap install serial-vault
```

The service needs a PostgreSQL database to run, so:
- Install PostgreSQL and create a database.
- Set up the config file, using ```settings.yaml``` as a guide.
- Configure the snap using the ```settings.yaml``` file:

```bash
$ cat /path/to/settings.yaml | sudo /snap/bin/serial-vault.config
$ sudo systemctl restart snap.serial-vault.service.service
```

The snap will create the tables in the database on restart, as soon as it has a valid database connection.

The service mode (signing or admin) is defined in the settings.yaml file. The
selected service should be accessible on port :8080 or :8081:
 - Signing Service: http://localhost:8080/v1/version
 - Admin Service: http://localhost:8081/
 - System-User Service: http://localhost:8082/

The Admin and System-User services' CSRF protection sends a cookie over a secure channel. If the cookie is to be sent
over an insecure channel, it is needed to workaround it by setting the environment variable:
```bash
$ export CSRF_SECURE=disable
```
When modified that environment variable is set, consider that the current web session must be invalidated
in order to changes take effect. That could require a browser restart.
NEVER set this configuration in production environments.

## Install from Source
If you have a Go development environment set up, Go get it:

  ```bash
  $ go get github.com/CanonicalLtd/serial-vault
  ```

### Configure it:
- Install PostgreSQL and create a database.
- Set up the config file, using ```settings.yaml``` as a guide.
- Create the database tables:
  ```bash
  $ cd serial-vault
  $ go run cmd/serial-vault-admin/main.go database --config=/path/to/settings.yaml
  ```

### Run it:
  ```bash
  $ cd serial-vault
  $ go run cmd/serial-vault/main.go -config=/path/to/settings.yaml -mode=signing
  ```

The application has an admin service that can be run by using mode=admin.

## Deploy it with Juju
Juju greatly simplifies the deployment of the Serial Vault. A charm bundle is available
at the [charm store](https://jujucharms.com/u/jamesj/serial-vault-bundle/), which deploys
everything apart from the Apache front-end units. There is an example of using Juju in the
[Deployment Guidelines](docs/Deployment.md).

The Juju charm uses a snap that is available at the [Snap store](https://uappexplorer.com/app/serial-vault.james)

## Try with docker
  ```bash
  $ git clone https://github.com/CanonicalLtd/serial-vault
  $ cd serial-vault/docker-compose
  $ docker-compose up
  # remove containers after try
  $ docker-compose kill && docker-compose rm
  ```

## Development Environment

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
nvm install v4.4.3

# Select the version to use
nvm ls
nvm use v4.4.3
```

- Install the nodejs dependencies
```bash
cd serial-vault
npm install
```

### Working with React

#### Build the project bundle
```bash
# Select the version to use
nvm ls
nvm use v4.4.3
npm run build
```

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
  "version":"0.1.0",
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


[travis-image]: https://travis-ci.org/CanonicalLtd/serial-vault.svg?branch=master
[travis-url]: https://travis-ci.org/CanonicalLtd/serial-vault
