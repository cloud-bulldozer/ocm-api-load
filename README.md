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
      --aws-access-key string      AWS access key
      --aws-access-secret string   AWS access secret
      --aws-account-id string      AWS Account ID, is the 12-digit account number.
      --aws-region string          AWS region (default "us-west-1")
      --config-file string         config file (default "config.yaml")
      --cooldown int               Cooldown time between tests in seconds. (default 10)
      --duration int               Duration of each individual run in minutes. (default 1)
      --end-rate int               Ending request per second rate. (E.g.: 5 would be 5 req/s)
      --gateway-url string         Gateway url to perform the test against (default "https://api.integration.openshift.com")
  -h, --help                       help for ocm-api-load
      --ocm-token string           OCM Authorization token
      --ocm-token-url string       Token URL (default "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
      --output-path string         Output directory for result and report files (default "results")
      --ramp-duration int          Duration of ramp in minutes, before normal execution. (default 0)
      --ramp-steps int             Number of stepts to get from start rate to end rate. (Minimum 2 steps)
      --ramp-type string           Type of ramp to use for all tests. (linear, exponential)
      --rate string                Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
      --start-rate int             Starting request per second rate. (E.g.: 5 would be 5 req/s)
      --test-id string             Unique ID to identify the test run. UUID is recommended (default "c160dab1-7fa3-4965-9797-47da16e5c1b9")
      --test-names strings         Names for the tests to be run.
  -v, --verbose                    set this flag to activate verbose logging.
```

## Tests

| Name | Endpoint | Method |
|----|----|----|
| self-access-token | /api/accounts_mgmt/v1/access_token | POST |
| list-subscriptions | /api/accounts_mgmt/v1/subscriptions | GET |
| access-review | /api/authorizations/v1/access_review | POST |
| register-new-cluster | /api/accounts_mgmt/v1/cluster_registrations | POST |
| register-existing-cluster | /api/accounts_mgmt/v1/cluster_registrations | POST |
| create-cluster | /api/clusters_mgmt/v1/clusters | POST |
| list-clusters | /api/clusters_mgmt/v1/clusters | GET |
| get-current-account | /api/accounts_mgmt/v1/current_account | GET |
| quota-cost | /api/accounts_mgmt/v1/organizations/{orgId}/quota_cost | GET |
| resource-review | /api/authorizations/v1/resource_review | POST |
| cluster-authorizations | /api/accounts_mgmt/v1/cluster_authorizations | POST |
| self-terms-review | /api/authorizations/v1/self_terms_review | POST |
| certificates | /api/accounts_mgmt/v1/certificates | POST |
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
- cooldown: Cooldown time between tests in seconds. (default 10 s)
- rate: Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
- test-id: Unique ID to identify the test run. UUID is recommended (default "dc049b1d-92b4-420c-9eb7-34f30229ef46")
- ramp-type: Type of ramp to use for all tests. (linear, exponential)
- ramp-duration: Duration of ramp in minutes, before normal execution. (default 0)
- start-rate: Starting request per second rate. (E.g.: 5 would be 5 req/s)
- end-rate: Ending request per second rate. (E.g.: 5 would be 5 req/s)
- ramp-steps: Number of stepts to get from start rate to end rate. (Minimum 2 steps)
- tests: List of the tests to run. Empty list means all.

### Test options

Each test can contain this options:

- rate: Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h') (default "1/s")
- duration: Override duration for the test. (A positive integer accompanied of a valid unit)

#### Ramping functionality

Each test can have a specific configuration for ranmping up the rate, inthis case the following options must be provided.

- duration: in minutes
- ramp-type: Type of ramp to use for all tests. (linear, exponential)
- ramp-duration: Duration of ramp in minutes, before normal execution. (default 0)
- start-rate: Starting request per second rate. (E.g.: 5 would be 5 req/s)
- end-rate: Ending request per second rate. (E.g.: 5 would be 5 req/s)
- ramp-steps: Number of stepts to get from start rate to end rate. (Minimum 2 steps)

> `rate` option is not needed for this.

##### Example

```yaml
  cluster-authorizations:
    duration: 30
    ramp-type: exponential
    ramp-duration: 10
    start-rate: 1
    end-rate: 50
    ramp-steps: 6
```

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

### Requirements

#### External

`vegeta` executable is necessary

`$ go get -u github.com/tsenart/vegeta`

#### python requirements

```bash
$ python3 -m venv env
$ . ./env/bin/activate
$ pip3 install -r requirements.txt
```

### Usage

To generate the report run the following command:

`python3 automation.py --dir /tests/2021-05-18`

The first argument should be the path to the folder where the `results` folder is located.

### Graph a specific file

`python3 automation.py graph --dir /tests/2021-05-18/results/ --filename access_review.json`

This should open the browser with an interactive Graph for access review.

### Generate `vegeta` reports

`python3 automation.py report --dir /tests/2021-05-18`

This will generate all the `vegeta` report files for each result file

When done deactivate virtual environment

```bash
$ deactivate
```

## How to release

Steps:

- Add your GitHub token to the env variable `GITHUB_TOKEN`

- Make sure you have [`github-release`](https://github.com/github-release/github-release#how-to-install) installed

- Be sure you are in the latest version of `main` branch and have bumped the version

- Now you are ready to run `make release` this will build the binary and generate the tarfiles that contain all the needed files

## Ramping Up Theory

The test will run a number <ramp-steps> with a running time of <duration>/<ramp-steps> rounded for each step, this can sometimes make the test last more or less than the expected duration, but we want to have a even distribution of times.

As each step finishes it will increase the rate according to a delta that is calculated with the parameters:

For both types of ramps we have common behaviour:

- First rate: is always `start-rate`
- Last rate: is always `end-rate`
- Since we cannot use float values for rates, we round all the rates to it's closest integer.

### For a linear ramp it will use this formula

`delta = ( end-rate - start-rate ) / ( ramp-steps - 1 )`
>ramp-steps, has always have to be greater than 1

So the new rate will be:

`newRate = oldRate + delta`

### For an exponential distribution

We are using the exponential formula `f(t)= x * <coeff> ^ t`

the `coeff` is calculated with this formula

`coeff = (end-rate / start-rate) ^ (1 / ramp-steps)`

So the new rate will be:

`newRate = start-rate * coeff ^ <# of step>`

### `duration` vs `ramp-duration`

The `duration` is the number of minutes the test is going to run. The `ramp-duration` is the number of minutes the ramp is going to last.

- If `ramp-duration` is not set, the ramp will take the whole `duration`.
- If `ramp-duration` is set, it will run the ramp for that long and then run the remaining of the `duration` at the `end-rate`.
  - E.g.: if `duration` is 30 minutes and `ramp-duration` is 20 minutes. The test will run a ramp for 20 minutes and keep running at `end-rate` for the remaining 10 minutes. So it will run `end-rate` for `duration` - `ramp-duration` minutes.
- If `ramp-duration` is greater than `duration` it will just perform a ramp for `ramp-duration` minutes.

Overrides for the values work the same, localized test values take priority over global values.
