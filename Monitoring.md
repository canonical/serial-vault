# Serial Vault Monitoring

Serial Vault server exposes monitoring data over endpoints for both services: admin and signing.

## Healthcheck Probe

Liveness probe is available at `/_status/ping`/. 
It will respond to the HTTP GET request with `200 OK` 
status and the version of the service in the response body.

Readiness probe is available at `/_status/check`. 
It will respond to the HTTP GET request with `200 OK`
status and json string in the body: `{"database": "OK"}`. On each request, the database connection will be
checked. 
On any error the endpoint will respond with `500  Internal Server Error` status with the 
json string in the body: `{"database": "database error string"}`


## Prometheus Probe

Prometheus data are available at `/_status/metrics`

### List of metrics exposed by Serial Vault

- standard go runtime metrics prefixed by `go_`
- process level metrics prefixed with `process_`
- prometheus scrape metrics prefixed with `promhttp_`
- Service-specific metrics for the incoming http requests are prefixed with `http_in`
  - `http_in_requests` metric for incoming HTTP requests count
  - `http_in_latency` metric for incoming requests latency in milliseconds
  - `http_in_errors` metric for HTTP errors (server responded with `5xx`)
  - `http_in_timeouts` metric for incoming HTTP timeouts (server responded with `504`)
    
