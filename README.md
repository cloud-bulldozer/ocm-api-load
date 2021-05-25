# OCM API LOAD TEST

![Go Version](https://img.shields.io/badge/go%20version-%3E=1.15-61CFDD.svg?style=flat-square)
![Python Version](https://img.shields.io/badge/python%20version-%3E=3.7-61CFDD.svg?style=flat-square&color=brightgreen)

A set of load tests for [OCM](https://github.com/openshift-online/ocm-api-model)'s clusters-service, based on vegeta.

## Requirements

- Go >= 1.15

To get all modules to local cache run

`go mod dowload`

## How to run?

Compile using `make` and run as a simple binary:

```sh
./ocm-load-test --test-id=foo --ocm-token=$OCM_TOKEN --duration=20m --rate=5/s --output-path=./results --test-names="<test_name>[,...]"

```

```sh
./ocm-load-test --config-file /path/to/config
```

if your `config-file` is named `config.yaml` and it in the same
path as your binary, it will be autodetected and you could run by just calling it:

```sh
./ocm-load-test
```

### Flags

> Note: Flags always take precedence over config file.
>> Default values don't count for precedence.

```
      --config-file string     config file (default "config.yaml")
      --duration int           Duration of each individual run in minutes. (default 1)
      --gateway-url string     Gateway url to perform the test against (default "https://api.integration.openshift.com")
  -h, --help                   help for ocm-api-load
      --ocm-token string       OCM Authorization token
      --ocm-token-url string   Token URL (default "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
      --output-path string     Output directory for result and report files (default "results")
      --rate string            Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
      --test-id string         Unique ID to identify the test run. UUID is recommended (default "dc049b1d-92b4-420c-9eb7-34f30229ef46")
      --test-names strings     Names for the tests to be run. (default [all])
```

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
| quota-cost | /api/accounts_mgmt/v1/organizations/{orgId}/quota_cost | GET |
| resource-review | /api/authorizations/v1/resource_review | POST |
| cluster-authorizations | /api/accounts_mgmt/v1/cluster_authorizations | POST |
|--|--|--|

## Config file

### Global options

- ocm-token: OCM Authorization token
- ocm-token-url: Token URL (default "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
- gateway-url: Gateway url to perform the test against (default "https://api.integration.openshift.com")
- client:
  - id: OpenID client identifier.
  - secret: OpenID client secret.
- output-path: Path to output results.
- duration: Duration of each individual run in minutes. (default 1)
- rate: Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
- test-id: Unique ID to identify the test run. UUID is recommended (default "dc049b1d-92b4-420c-9eb7-34f30229ef46")
- tests: List of the tests to run. Empty list means all.

### Test options

Each test can contain this options:

- rate: Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
- duration: Override duration for the test. (A positive integer accompanied of a valid unit)

### Obligatory options

- ocm-token
- gateway-url

### Defaults

- output-path: set to `./results`
- duration: set to `1` minute

### Minimal yaml config file

```yaml
---
token: xxxXXXyyyYYYzzzZZZ
gateway-url: https://api.my-env.openshift.com
```

### Full yaml config file

See [example](./config.example.yaml)

## Reporting

### Rquirements

#### External

`vegeta` executable is necessary

`$ go get -u github.com/tsenart/vegeta`

#### python requirements

`$ pip3 install -r requirements.txt`

### Usage

To generate the report run the following command:

`python3 export_report.py /tests/2021-05-18`

The first argument should be the path to the folder where the `results` folder is located.

### Graph a specific file

`python3 export_report.py graph /tests/2021-05-18/results/access_review.json`

This should open the browser with an interactive Graph for access review.

### Generate `vegeta` reports

`python3 export_report.py report /tests/2021-05-18`

This will generate all the `vegeta` report files for each result file
