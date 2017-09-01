---
title: "/v1/serial"
table_of_contents: False
---

## POST /v1/serial

### Description

Generate a serial assertion signed by the brand key.
Takes the details from the device as a serial-request assertion and generates a signed serial assertion

### Request

The message must be the serial-request assertion format and is best generated using
snapd libraries.

Main fields in request are:

| Field | Description |
|-------|-------------|
| brand-id | the Account ID of the manufacturer (string) |
| model | the name of the device (string) |
| device-key | the encoded type and public key of the device (string) |
| request-id | the nonce returned from the /v1/request-id method previous call (string) |
| signature | the signed data |
| serial | serial number of the device (string)|


### Response

The method returns a signed serial assertion using the key from the vault.
see details [here](https://docs.ubuntu.com/core/en/reference/assertions/serial)

### Errors

The following errors can occur:

 * Error in retrieving the authentication token
 * The authentication token is invalid
 * Error encoding the version response

### Example

```
POST /v1/serial HTTP/1.1
Host: serial-vault
Accept: image/gif, image/jpeg, */*
Accept-Language: en-us
Accept-Encoding: gzip, deflate
User-Agent: Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)

type: serial-request
brand-id: System
mode: pc-amd64
device-key:
    AcbBTQRWhcGAARAAx6VJoV9ZKASKa1pFA0G6hQimQT7ym8EZFN7+SzZhWSWLIwFd06oRQVKetQB6
    a+ab0zMN3yfI94aB9aH/q6vA7T7Yo1KaBFy4aaztUvDmMzEGaVwJvDSBUBFr4yUCJEtLXAw5fMkS
    DGvNUFRacLifAfGU5mLHJl7WXY2e7T+VjJPoSU3nAZjvGd2YQnQ1fNfQ0X+zuQVDGrtmJJF3x0CM
    8LL0XF4UCTBYyLZK2YvSKrrk2qmIUVr3PXoY+fH9Bs5AZAAZ91GIrt0qc0uradXxI6kq8zy8bVl8
    GTazEmkBE9Y7snAqWJWGXt9K4tO7h+4Xgprvf27dddp68XS2KHT3r86qC/1i9mTGMbHWJ5NKd/No
    Jnawjc1qo2tnVVyw+GKwMhukpvmtuejhtk395dNczGZ2sw2yPHORUHUyq/sPLoAWyWLQFHL3MxQq
    qyxgxWNnRYhcs6wmWEf2nNFlllld6YzS7It+cA+I04j5h85DGO6+knn1J7X4WuORDx3nn3bEQKik
    v4uu1xFJYk6N14B/ofMoUCzbPtgkNpmV0NmgFeogx+I5yRuF0EF5U+LfMuAE+ROoYHHwiBHeSttr
    YewdunntDyeRUc3CTwsvfq2zARObr5He5z4ldSASuzxbzEEXVd6UERPN+zeJGyctKIYEqvpSNNuu
    4Fs8Ctp6yar9KucAEQEAAQ==
request-id: REQID
body-length: 10
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

AcLBUgQAAQoABgUCV7R2CQAAYx4QAB2vxjMYFb5nmQdkeX4pjbD6sjheD5PZV6h3DDDznZccMP+v
y3x8PtTA7h1oN04nzMBqilPH01buSVSSzVAy789oecAwSMhpUi50lVIWdye2zeE+G3DEbZdHOBod
+rxK0LTuDxf8dCCt2zbGlm/4wSORGPsn4dR+G6Da+ZEEAORQuHCdVGNe9LgFi7ZIX5ZkvK5oNTyH
Ebgf4VLVpHpBZ2sl6sNPwLDpH1LOmMFgq3tEZXaKaa9QAn6g/S/hgTbv6eDfKHTX99ynpqgu6+am
+HZ28PG39kbJoKpexzIxhxhR42hKso3xUHJfwFSeTxLIlRK0KlDRsDOAe6MzjhTnA8b/xMjw8NaF
6q60hgS8Qytyvu1/7f75CTy4cTwenmUuw/v2mcO98FurVpDFzXSb5HK44Ej6gYXpTOtE4lSH0oP4
7VL/JAjhP3qncgDMVh0URIqh6FDCD7bb2USP4Fo2yvkVfLHCS80vZGury+rGxV2bPRcOTfbnoKZy
cwmwjJS6vKEYIIlMwVaHsPd9ZBvyYBwTzfGKtoazjm44mByBG0AEUZrZ7MWnf7lWwU+Ze3g3GNQF
9EEnrN8E9yYxFgCGaYA7kBFhkhJElafMQNr/EYU3bwLKHa++1iKmNKcGePRZyy9kyUpmgtRaQt/6
ic2Xx1ds+umMC5AHW9wZAWNPDI/T


HTTP/1.1 200 OK
Date: Sun, 18 Oct 2009 08:56:53 GMT
Server: Apache/2.2.14 (Win32)
Last-Modified: Sat, 20 Nov 2004 07:16:26 GMT
ETag: "10000000565a5-2c-3e94b66c2e680"
Accept-Ranges: bytes
Content-Length: 44
Connection: close

type: serial
authority-id: canonicañ
brand-id: System
model: pc-amd64
serial: 03961d5d-26e5-443f-838d-6db046126bea
device-key:
    AcbBTQRWhcGAARAA0y/BXkBJjPOl24qPKOZWy7H+6+piDPtyKIGfU9TDDrFjFnv3R8EMTz1WNW8d
    5nLR8gjDXNh3z7dLIbSPeC54bvQ7LlaO2VYICGdzHT5+68Rod9h5NYdTKgaWDyHdm2K1v2oOzmMF
    Z+MmL15TvP9lX1U8OIVkmHhCO7FeDGsPlsTX2Wz++SrOqG4PsvpYsaYUTHE+oZ+Eo8oySW/OxTmp
    rQIEUoDEWNbFR5/+33tHRDxKSjeErCVuVetZxlZW/gpCx5tmCyAcBgKoEKsPqrgzW4wUAONaSOGc
    Zuo35DxwqeGHOx3C118rYrGvqA2mCn3fFz/mqnciK3JzLemLjw4HyVd1DyaKUgGjR6VYBcadL72n
    YN6gPiMMmlaAPtkdFIkqIp1OpvUFEEEHwNI88klM/N8+t3JE8cFpG6n4WBdHUAwtMmmVxXm5IsM3
    uNwrZdIBUu4WOAAgu2ZioeHLIQlDGw6dvVTaK+dTe0EXo5j+mH5DFnn0W1L7IAj6rX8HdiM5X5fP
    4kwiezSfYXJgctdi0gizdGB7wcH0/JynaXA/tI3fEVDu45X7dA/XnCEzYkBxpidNfDkmXxSWt5N/
    NMuHZqqmNHNfLeKAo1yQ/SH702nth6vJYJaIX4Pgv5cVrX5L429U5SHV+8HaE0lPCfFo/rKRJa9i
    rvnJ5OGR4TeRTLsAEQEAAQ==
device-key-sha3-384: _4U3nReiiIMIaHcl6zSdRzcu75Tz37FW8b7NHhxXjNaPaZzyGooMFqur0EFCLS6V
timestamp: 2016-11-08T18:16:12.977431Z
sign-key-sha3-384: BWDEoaqyr25nF5SNCvEv2v7QnM9QsfCc0PBMYD_i2NGSQ32EF2d4D0hqUel3m8ul

AcLBUgQAAQoABgUCWCIWcgAARegQAB4/UsBpzqLOYOpmR/j9BX5XNyEWxOWgFg5QLaY+0bIz/nbU
avFH4EwV7YKQxX5nGmt7vfFoUPsRrWO4E6RtXQ1x5kYr8sSltLIYEkUjHO7sqB6gzomQYkMnS2fI
xOZwJs9ev2sCnqr9pwPC8MDS5KW5iwXYvdBP1CIwNfQO48Ys8SC9MdYH0t3DbnuG/w+EceOIyI3o
ilkB427DiueGwlBpjNRSE4B8cvglXW9rcYW72bnNs1DSnCq8tNHHybBtOYm/Y/jmk7UGXwqYUGQQ
Iwu1W+SgloJdXLLgM80bPzLy+cYiIe1W1FSMzVdOforTkG5mVFHTL/0l4eceWequfcxU3DW9ggcN
YJx8MPW9ab5gPibx8FeVb6cMWEvm8S7wXIRSff/bkHMhpjAagp+A6dyYsuUwPXFxCvHSpT0vUwFS
CCPHkPUwj54GjKAGEkKMx+s0psQ3V+fcZgW5TBxk/+J83S/+6AiQ06W8rkabWCRyl2fX81vMBynQ
nu147uRGWTXfa31Mys9lAGNHMtEcMmA106f2XfATqNK99GlIIjOxqEe5zH3j51JtY+5kyJd9cqvl
Pb0rZnPySeGxnV4Q2403As67AJrIExRrcrK2yXZjEW3G2zTsFNzBSSZr0U8id1UJ/EZLB/em2EHw
D2FXTwfDiwGroHYUFAEu1DkHx7Sy
}
```