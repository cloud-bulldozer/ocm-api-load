# OCM API LOAD TEST

A set of load tests for OCM's clusters-service, based on vegeta.

## How to run?

Compile using `make` and run as a simple binary:

```sh
./ocm-api-load --config-file /path/to/config
```

if your `config-file` is named `ocm-api-load.yaml` and it in the same
path as your binary, it will be autodetected and you could run by just calling it:

```sh
./ocm-api-load
```

### Command options

- config-file: Path to the configuration file.
- token-url: URL to obtain a token.

## Tests

| Name | Endpoint | Method |
|----|----|----|
| self-access-token | /api/accounts_mgmt/v1/access_token | POST |
| list-subscriptions | /api/accounts_mgmt/v1/subscriptions | GET |
| access-review | /api/authorizations/v1/access_review | POST |
| register-new-cluster | /api/accounts_mgmt/v1/cluster_registrations | POST |
| create-cluster | /api/clusters_mgmt/v1/clusters | POST |
| list-clusters | /api/clusters_mgmt/v1/clusters | GET |
| get-current-account | /api/accounts_mgmt/v1/current_account | GET |
|--|--|--|

## Config file

### Global options

- token: Offline token for authentication.
- gateway-url: Gateway URL.
- client:
  - id: OpenID client identifier.
  - secret: OpenID client secret.
- duration-minutes: How long should the load test take.
- output-path: Path to output results.

### Test options

Each test can contain this options:

- freq: Number of requests to execute in a unit of time.
- per: Unit of the request frequency. ("ns", "us" (or "µs"), "ms", "s", "m", "h")
- duration: Override duration for the test. (A positive integer accompanied of a valid unit)

### Obligatory options

- token
- gateway-url

### Defaults

- output-path: set to `./results`
- duration-minutes: set to `1` minute

### Minimal yaml config file

```yaml
---
token: xxxXXXyyyYYYzzzZZZ
gateway-url: https://api.my-env.openshift.com
```

### Full yaml config file

```yaml
---
token: xxxXXXyyyYYYzzzZZZ           # Offline token for authentication.
gateway-url: http://localhost:8000  # Gateway URL.
client:
  - id: cloud-services              # OpenID client identifier.
  - secret: "secure-secret"         # OpenID client secret.
duration-minutes: 1                 # How long should the load test take.
output-path: "./results"            # path to output results.
tests:
  self-access-token:
    freq: 1000
    per: "h"
  list-subscriptions:
    freq: 2000
    per: "h"
  access-review:
    freq: 100
    per: "s"
  register-new-cluster:
    freq: 1000
    per: "h"
  create-cluster:
    freq: 10
    per: "s"
  list-clusters:
    freq: 10
    per: "s"
    duration: "20s"
  get-current-account:
    freq: 6
    per: "m"
```
