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
  * It is common to request reviews from those reviewers automatically suggested by GitHub
* Each PR needs to have been open for at least 24 working hours to allow for community feedback
  * The 24 working hours counts hours occurring Mon-Fri in the local timezone of the submitter
* Each PR must be fully updated to master and tests must have passed

When the criteria are met, a project maintainer can merge your changes into the project's master branch.

## Troubleshooting and Debugging

**Dumping etcd key and values:**

Download `etcdctl` binary from [etcd releases](https://github.com/coreos/etcd/releases).
Run the following command to dump all keys and values in etcd to stdout. Replace
`endpoints` argument to point to etcd client URL if etcd is running on a
different ip and port.

```sh
ETCDCTL_API=3 etcdctl get --prefix=true "" --endpoints="127.0.0.1:2379"
```

**Generating REST API documentation:**

```sh
$ curl -o endpoints.json -s -X GET http://127.0.0.1:24007/endpoints
$ go build pkg/tools/generate-doc.go
$ ./generate-doc
```

You should commit the generated file `doc/endpoints.md`

**Setup tracing:**

Tracing glusterd2 operations is accomplished using [OpenCensus Go](https://github.com/census-instrumentation/opencensus-go), which is a Go implementation of OpenCensus. The tracing implementation uses [Jaeger](https://www.jaegertracing.io/) as the backend to export tracing data. The Jaeger UI can then be used to visualize the captured traces.

Run the following steps to setup and view tracing using Jaeger. Note that the steps outlined below is a quick way to setup and view traces for debugging GD2 without attaching a backing store to the Jaeger service.

1. Prior to starting GD2 on any node, start the Jaeger service either on your local machine or on a server/VM. The Jaeger service can be started either as a standalone service or within a container using an available docker image.

  * **Standalone service:** See the "Running Individual Jaeger Components" section in the [Getting Started](https://www.jaegertracing.io/docs/getting-started/) page of Jaeger documentation.

  * **Docker Image:** For quick local testing, an all-in-one docker image can be used which launches the Jaeger UI, query and agent. This image comes with an in-memory storage component. For example, the following command starts the all-in-one docker image with the required services,
  ```sh
  $ docker run -d -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:latest
  ```
  >NOTE: The Jaeger agent runs on port 6831/6832. The Jaeger collector runs on port 14268. The Jaeger query service runs on port 16686 on which the UI can be accessed. Ensure that firewalld is configured (or stopped) to let traffic on the Jaeger specific ports.

2. Start GD2 process on the gluster nodes and provide the Jaeger endpoints either within the config file using the `--config` option or pass the endpoints as separate options. For the Jaeger service to capture traces, the Jaeger agent endpoint  and the Jaeger collector endpoint are necessary. The following outlines both the ways,

 * **Config File:** Add the following to your startup config (for e.g. conf.toml) file on each node.
  ```toml
  ...(existing options)
  jaeger-endpoint = "http://192.168.122.1:14268"
  jaeger-agent-endpoint = "http://192.168.122.1:6831"
  ```
  Start GD2 as usual on each node. For e.g.,
  ```sh
  $./glusterd2 --config conf.toml
  ```
  >NOTE: Change the IP address based on your configuration.

 * **Start-up Option:** Provide the Jaeger endpoints as options to glusterd2 start-up command in case you don't wish to provide it in the config file. For e.g., the example below shows the options passed to glusterd2. NOTE: Provide the options for all the nodes in your gluster cluster.
  ```sh
  $./glusterd2 --config conf.toml --jaeger-endpoint http://192.168.122.1:14268 --jaeger-agent-endpoint http://192.168.122.1:6831
  ```

3. Verify that on start-up, GD2 was successfully able to connect to the Jaeger endpoints by looking for the following GD2 start-up log message,
  ```log
  ...
  INFO[2018-07-25 13:24:55.174171] tracing: Registered opencensus jaeger exporter for traces and stats  jaegerAgentEndpoint="http://192.168.122.1:6831" jaegerEndpoint="http://192.168.122.1:14268" source="[tracing.go:67:tracing.InitJaegerExporter]"
  ...
  ```
  >NOTE: In case of any warning or error message, verify firewalld settings and the status of Jaeger services.

4. Execute the intended GD2 operation (for e.g. volume create) and view the traces on the Jaeger UI by navigating to the endpoint. For e.g. if the Jaeger service was started locally, then navigate to `http://localhost:16686`. An example of how a trace looks like for a replica 3 volume create transaction is shown in this [github issue](https://github.com/gluster/glusterd2/issues/1049).
