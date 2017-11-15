## New to Go?
Glusterd2 is written in Go and if you are new to the language, it is **highly** encouraged to:

* Take the [A Tour of Go](http://tour.golang.org/welcome/1) course.
* [Set up](https://golang.org/doc/code.html) Go development environment on your machine.
* Read [Effective Go](https://golang.org/doc/effective_go.html) for best practices.

## Development

### Workspace and repository setup

1. [Download](https://golang.org/dl/) Go (>=1.8) and [install](https://golang.org/doc/install) it on your system.
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

Do your work here and commit. You will need to provide unit tests and functional tests for your changes wherever applicable. Ensure that the tests pass with your changes:

`# make test`

Once you are ready to push, you will type the following:

`$ git push fork hellopatch`

**Creating A Pull Request:**   
When you are satisfied with your changes, you will then need to go to your repo in GitHub.com and create a pull request for your branch. Automated tests will be run against the pull request. Your pull request will be reviewed and merged.
