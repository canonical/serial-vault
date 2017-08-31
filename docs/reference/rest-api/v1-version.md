---
title: "/v1/version"
table_of_contents: False
---

## GET /v1/version

### Description

Returns the version of the Serial Vault web service

### Request

None

### Response

```
{
  "version":"0.1.0",
}
```

| Field | Description |
|---------------|-----|
| version  | the version of the serial vault service (string) |

### Errors

The following errors can occur:

 * Error in retrieving the authentication token
 * The authentication token is invalid
 * Error encoding the version response

### Example

```
wget https://serial-vault/v1/version
{
  "version":"0.1.0",
}
```