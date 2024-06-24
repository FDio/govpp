Integration Tests
=================

The GoVPP integration testing suite runs each test case against real VPP instance.

The integration tests cases are grouped by files:
- `binapi` - runs tests for VPP API
- `examples` - run examples as tests
- `stats` - runs tests for VPP Stats
- `trace` - runs tests for VPP API trace
- `*` (*other*)  - runs specialized tests

## Running Tests

The recommended way to run the integration tests is to use a self-contained testing environment, which is managed from a helper bash script [`run_integration.sh`](../run_integration.sh). The script will build a special Docker image that includes testing suite and other requirements (e.g. VPP, gotestsum..) and then will run the integration tests inside a container.

```shell
make test-integration
```

This will run the tests against latest VPP release by default. To run against specific VPP version, add `VPP_REPO=<REPO>` where `<REPO>` is name of VPP repostiroy on packagecloud.

```shell
# Run against specific VPP version
make test-integration VPP_REPO=2306

# Run against VPP master branch
make test-integration VPP_REPO=master
```

The make target above simply runs a helper script which accepts additional arguments that are passed down directly to `go test ...`.

```shell
# Usage:
#  ./test/integration/run_integration.sh <ARGS>

# Run with verbose mode
./test/run_integration.sh -test.v
```

### Run Specific Test Case

To run a specific integration test case(s):

```shell
./test/run_integration.sh -test.run="Interfaces"
```

## Running Tests on your Host

If the script `run_integration.sh` is not used to run tests and the test cases
are directly used, the tests will try to start/stop VPP instance for each test
case individually.

> **Warning**
> This method requires VPP to be installed on your host system.

```shell
TEST=integration go test ./test/integration
```
