---
title: "/v1/request-id"
table_of_contents: False
---

## POST /v1/request-id

### Description

Returns a nonce that is needed for the 'serial' request.

### Request

None. Though header must include model api-key
```
api-key: <the_api_key_value>
```
### Response

```
{
  "request-id": "abc123456",
  "success": true,
  "message": ""
}
```
| Field | Description |
|-------|-------------|
| request-id* | unique string that is needed for serial requests (string) |
| success* | whether the request was successful (bool) |
| message* | error message from the request (string) |

### Errors

The following errors can occur:

 * Error in retrieving the authentication token
 * The authentication token is invalid
 * Invalid API key used
 * delete-expired-nonces
 * generate-request-id error

### Example

```
wget --header='api-key: 47ladfh4la8009dafhYYZ0' https://serial-vault/v1/request-id
{
  "request-id": "abc123456",
  "success": true,
  "message": ""
}
```