# CS LOAD TEST

A set of load tests for OCM's clusters-service, based on vegeta.

## How to run?

Compile using `make` and run as a simple binary:

```sh
./cs-load-test --token=<OCM_TOKEN>   \
    --output-path="/path/to/output/" \
    --gateway-url=<api-gateway-url>  \
    --duration-in-min=1              \
```

## CLI Parameters

1. The `token` paramater is an auth token for OCM. One can obtain a token from the following [link](https://qaprodauth.cloud.redhat.com/openshift/token).
2. The `gateway-url` parameter is a url to run the tests against, by default `http://localhost:8000` which is the default endpoint for the development environment.
3. The `duration-in-min` parameter indicates how long should the test run in minutes.
4. The `output-path` paramater indicates the path to output the results. For more information about the results format, please see [vegeta](https://github.com/tsenart/vegeta#report-command).
