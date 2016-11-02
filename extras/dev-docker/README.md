# Docker image for GD2 development

This docker image contains everything required to build and run GD2 during development.
The image is available as a trusted build from Docker hub `gluster/glusterd2-dev`.

This image is used in the 'Vagrantfile' provided at the root of the GD2 repo to setup a development environment for GD2 development.

To build a local image run,
```
$ docker build -t gluster/glusterd2-dev:latest .
```
from this directory.

## History

This docker image was originally developed at [kshlm/glusterd2-dev-docker](https://github.com/kshlm/glusterd2-dev-docker),
and was available as `kshlm/glusterd2-dev` from Docker hub.
