## New to Go?
Glusterd2 is written in Go and if you are new to the language, it is **highly** encouraged to:

* Take the [A Tour of Go](http://tour.golang.org/welcome/1) course.
* [Set up](https://golang.org/doc/code.html) Go development environment on your machine.
* Read [Effective Go](https://golang.org/doc/effective_go.html) for best practices.

## Development Workflow

### Workspace and repository setup

1. [Download](https://golang.org/dl/) Go (>=1.9) and [install](https://golang.org/doc/install) it on your system.
1. Setup the [GOPATH](http://www.g33knotes.org/2014/07/60-second-count-down-to-go.html) environment.
1. Run `$ go get -d github.com/gluster/glusterd2`  
   This will just download the source and not build it. The downloaded source will be at `$GOPATH/src/github.com/gluster/glusterd2`
1. Fork the [glusterd2 repo](https://github.com/gluster/glusterd2) on Github.  
1. Add your fork as a git remote:  
   `$ git remote add fork https://github.com/<your-github-username>/glusterd2`
1. Run `$ ./scripts/install-reqs.sh`

>  Editors: Our favorite editor is vim with the [vim-go](https://github.com/fatih/vim-go) plugin, but there are many others like [vscode](https://github.com/Microsoft/vscode-go).

### Building Glusterd2

To build Glusterd2 run:  
`$ make`

The built binary will be present under `build/` directory.

or to install run:  
`$ make install`

The built binary will be installed under `$GOPATH/bin/` directory.

### Code contribution workflow

Glusterd2 repository currently follows GitHub's [Fork & Pull](https://help.github.com/articles/about-pull-requests/) workflow for code contributions.

Please read the [coding guidelines](coding.md) document before submitting a PR.

Here is a short guide on how to work on a new patch.  In this example, we will work on a patch called *hellopatch*:

1. `$ git checkout master`
1. `$ git pull`
1. `$ git checkout -b hellopatch`

Do your work here and commit.

Run the test suite, which includes linting checks, static code check, and unit
tests:

`$ make tests`

You will need to provide unit tests and functional tests for your changes
wherever applicable. Ensure that the tests pass with your changes. The
functional tests needs to be run as root user. To run the functional tests:

`# make functest`

Once you are ready to push, you will type the following:

`$ git push fork hellopatch`

**Creating A Pull Request:**   
When you are satisfied with your changes, you will then need to go to your repo in GitHub.com and create a pull request for your branch. Automated tests will be run against the pull request. Your pull request will be reviewed and merged.

If you are planning on making a large set of changes or a major architectural change it is often desirable to first build a consensus in an issue discussion and/or create an initial design doc PR. Once the design has been agreed upon one or more PRs implementing the plan can be made.

**Review Process:**
Once your PR has has been submitted for review the following critieria will need to be met before it will be merged:
* Each PR needs reviews accepting the change from at least two developers for merging
  * It is common to request reviews from those reviewers automatically suggested by github
* Each PR needs to have been open for at least 24 working hours to allow for community feedback
  * The 24 working hours counts hours occuring Mon-Fri in the local timezone of the submitter
* Each PR must be fully updated to master and tests must have passed

When the criteria are met a project maintainer can merge your changes into the project's master branch.

## Troubleshooting and Debugging

**Dumping etcd key and values:**

Download `etcdctl` binary from [etcd releases](https://github.com/coreos/etcd/releases).
Run the following command to dump all keys and values in etcd to stdout. Replace
`endpoints` argument to point to etcd client URL if etcd is running on a
different ip and port.

```sh
ETCDCTL_API=3 etcdctl get --prefix=true "" --endpoints="127.0.0.1:2379"
```

**Connecting to external etcd cluster:**

Edit **glusterd2.toml** config option and add `noembed` option with specifying
the etcd endpoint:

```toml
etcdendpoints = "http://127.0.0.1:2379"
noembed = true
```

**Generating REST API documentation:**

```sh
$ curl -o endpoints.json -s -X GET http://127.0.0.1:24007/endpoints
$ go build pkg/tools/generate-doc.go
$ ./generate-doc
```

You should commit the generated file `doc/endpoints.md`
