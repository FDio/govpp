Integration Tests
=================

The GoVPP integration testing suite runs each test case against real VPP instance.

The tests cases are grouped in the following structure:
- `binapi` - runs tests for VPP API
- `examples` - run examples as tests
- `stats` - runs tests for VPP Stats
- `trace` - runs tests for VPP API trace
- `*` (*other*)  - runs specialized tests


## Running Tests

### Test in Docker container

The recommended way to run the integration tests is by using a self-contained environment.
There is a bash script that will build a Docker container for the test suite and
run the tests inside that container.

### Run Entire Test Suite

To run entire integration test suite:

```shell
./test/integration/run_integration.sh
```

### Run Specific Test Case

To run a specific integration test case(s):

```shell
./test/integration/run_integration.sh -test.run="Interfaces"
```

### Test on the host

If the script `run_integration.sh` is not used to run tests and the test cases
are directly used, the tests will try to start/stop VPP instance for each test
case individually.

> **Warning**
> This method requires VPP to be installed on the host system

```shell
go test -v ./test/integration
```
