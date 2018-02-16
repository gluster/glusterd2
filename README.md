# GlusterD-2.0

[![Go Report Card](https://goreportcard.com/badge/github.com/gluster/glusterd2)](https://goreportcard.com/report/github.com/gluster/glusterd2)
[![Build Status](https://ci.centos.org/view/Gluster/job/gluster_glusterd2/badge/icon)](https://ci.centos.org/view/Gluster/job/gluster_glusterd2/)

GlusterD-2.0 (GD2) is a re-implementation of GlusterD. It attempts to have better
consistency, scalability and performance when compared with the current
GlusterD, while also becoming more modular and easing extensibility.

## Documentation

* [Quick Start User Guide](doc/quick-start-user-guide.md)
* [Development Guide](doc/development-guide.md)
* [Coding Guidelines](doc/coding.md)
* [REST API Reference](doc/endpoints.md)

## Architecture and Design
Please refer to the [wiki](https://github.com/gluster/glusterd2/wiki/Design) for more information.

## Building

To build GD2, just run `make`. If you don't have the required tools installed, run `scripts/install-reqs.sh`.

## Contributing

We use the Github pull-request model for accepting contributions. If you are not familiar with the pull request model please read ["Using pull requests"](https://help.github.com/articles/using-pull-requests/). For specific information on GlusterD-2.0, refer the [Development Guide](doc/development-guide.md).

## Copyright and License
Copyright (c) 2015 Red Hat, Inc. <http://www.redhat.com>

This program is free software; you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License, version 3 or any later version (LGPLv3 or later), or the GNU General Public License, version 2 (GPLv2), in all cases as published by the Free Software Foundation.

