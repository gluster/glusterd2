# Testing

## Unit Tests

Write tests in files whose name ends with `_test.go` which ensures that they
are picked up by `go test` command. These test files will contain functions
with function name matching `TestXxx` pattern. Put these test files in the
same package as the one being tested. This ensures that unexported functions
can also be unit tested.

A test is not a unit test if it is not testing a single unit (a single
function or operation). A unit test should not bring up processes, make
actual network calls or use local filesystem (except /tmp). Unit tests
should ideally be idempotent.

Refer to documentation of go's [testing](https://golang.org/pkg/testing/)
package for detailed information.

**Running unit tests:**

```sh
$ make test
```

## Functional Tests

Functional tests (a.k.a black-box testing) are to be placed in the `e2e`
package in the root of the source repo. `e2e` stands for end to end testing
(borrowed from etcd and kubernetes). These tests will interact with live
instances of glusterd2 process or cluster just like how a user would. These
tests should not import any of glusterd's packages or make assumptions about
implementation details.

Functional tests are also run by the `go test` command but are disabled by
default as they can consume resources and change the environment they are
running in. You can pass `-functest` to `go test` command to run functional
tests.

**Running functional tests:**
```sh
# go test ./e2e -v -functest
```

**Running a single functional test:**
```sh
# go test ./e2e -v -functest -run "TestVolume"
```

The argument to the -run command-line flag is an unanchored regular expression
that matches the test's name.

## Retriggering tests in the CentOS CI system

The CentOS CI system runs `e2e` tests for every proposed PR and reports the
result back in the PR with a link to output of the tests.

If you are certain that the tests run by CentOS CI have failed spuriously, you
can retrigger running the `e2e` tests by writing the following magic phrase
on the PR as a comment:

```
retest this please
```
