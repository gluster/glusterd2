This document describes how to setup a development environment for _GlusterD-2.0_.

## Requirements
- [Go](http://golang.org/)
- [Docker](https://www.docker.com/)
- [Vagrant](http://www.vagrantup.com/)

### Installation instructions for Fedora 20

- Install Go and Docker

  ```
  yum install golang docker-io
  ```

- Install Vagrant

  ```
  yum install https://dl.bintray.com/mitchellh/vagrant/vagrant_1.6.5_x86_64.rpm
  ```

  > *NOTE*
  >
  > Get the latest Vagrant RPM link from http://www.vagrantup.com/downloads.html

## Setting up the development environment

- First create a directory for use as the GOPATH for glusterd-2.0 development

  ```
  mkdir -p $HOME/glusterd2-dev
  ```

- Export this path as GOPATH

  ```
  export GOPATH=$HOME/glusterd2-dev
  ```

- Get the sources and install consul and glusterd2

  ```
  go get github.com/hashicorp/consul
  go get github.com/kshlm/glusterd2
  ```

- Copy the Vagrantfile available with the glusterd2 package into GOPATH

  ```
  cp $GOPATH/src/github.com/kshlm/glusterd2/Vagrantfile $GOPATH
  ```

- Change into GOPATH and start the containers

  ```
  cd $GOPATH
  vagrant up
  ```

  This will startup 4 containers called glusterd-dev-{1..4}. The GOPATH from host will be attached to these containers as volumes. The containers have been setup to use the volumes as their local GOPATHs. This allows you the do changes to the source and compile on the host, and have the results immediately available in the containers for testing.

  - The first time you start the containers use the `--no-parallel` option to `vagrant up` to allow Docker to safely download the initial container image

    ```
    vagrant up --no-parallel
    ```

  - If the containers fail to startup with an error along the lines 'You need to specify the box image to use', you need to specify the provider to be used by vagrant

    ```
    vagrant --provider=docker up
    ```

- You can now SSH into the containers and begin your development/testing etc.

  ```
  vagrant ssh <container-name>
  ```

