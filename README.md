[![Build Status][travis-image]][travis-url] [![Coverage Status][coveralls-image]][coveralls-url]
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
  $ go run server.go -config=/path/to/settings.yaml
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

```json
{
  "serial":"M12345/LN",
  "brand-id": "System",
  "model":"Device 1000",
  "device-key":"ssh-rsa abcd1234",
  "revision": 2
}
```
- brand-id: the Account ID of the manufacturer (string)
- model: the name of the device (string)
- serial: serial number of the device (string)
- device-key: the type and public key of the device (string)
- revision: the revision of the device (integer)

#### Output message

```json
{
  "success":true,
  "message":"",
  "signature":"-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA256\n\ntype: device\nbrand-id: System\nmodel: Device 1000\nserial: M12345/LN\ntimestamp: 2016-02-03 17:22:59.93489652 +0000 UTC\nrevision: 2\ndevice-key: ssh-rsa abcd1234\n-----BEGIN PGP SIGNATURE-----\n\nwsFcBAEBCA ... A5LT\n-----END PGP SIGNATURE-----"}
```
- success: whether the submission was successful (bool)
- message: error message from the submission (string)
- identity: the formatted, clear-signed data (string)

#### Example
```bash
curl -X POST -d '{"serial":"M12345/LN", "brand-id":"System", "model":"Device 1000", "revision": 2, "device-key":"ssh-rsa abcd1234"}' http://localhost:8080/1.0/sign
```
