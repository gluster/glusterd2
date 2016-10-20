# GD2 test environment - Docker + Vagrant

This directory contains a `Dockerfile` to build a docker image with GD2 installed.
A 'Vagrantfile' is provided which makes use of this docker image to setup a test env.

## Docker image

A trusted build 'gluster/glusterd2-test' is available from the Docker hub.

To build the image on your own, run the build script from this directory
```
$ ./build.sh
```

The image has GD2 installed at `/usr/sbin/glusterd2`.

The `Dockerfile` and image will be update with every development release of GD2.

## Vagrant

The `Vagrantfile` sets up 4 running containers with GD2 installed.
Start the environment by running
```
$ vagrant up --provider docker
```

This brings up 4 containers named `gd2-{1..4}`
You can now SSH into the containers with,
```
$ vagrant ssh <name>
```

To stop the containers run,
```
$ vagrant halt
```

To destroy,
```
$ vagrant destroy -f
```
