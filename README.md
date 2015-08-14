# GlusterD-2.0
[![wercker status](https://app.wercker.com/status/6b02e386a99296c01e87cb8293fe3dc4/m/master "wercker status")](https://app.wercker.com/project/bykey/6b02e386a99296c01e87cb8293fe3dc4)

GlusterD-2.0 is a re-implementation of GlusterD. It attempts to be have better
consistency, scalability and performance when compared with the current
GlusterD, while also becoming more modular and easing extensibility.

## Architecture and Design
> NOTE: This is still under discussion. We will add details on this soon.

## Contributing

We are using [GerritHub](https://review.gerrithub.io) to review and accept changes to this repository.
The development process involves the following steps.

### Setting up

0. Register on [GerritHub](https://review.gerrithub.io) and get the ssh git address from https://review.gerrithub.io/#/admin/projects/kshlm/glusterd2

1. Clone this repository using `go get`
```
$ go get github.com/kshlm/glusterd2
```

2. Switch to the repository inside your `$GOPATH`
```
$ cd $GOPATH/src/github.com/kshlm/glusterd2
```

3. Add the ssh git repo address as a remote named `gerrit`
```
$ git remote add gerrit <ssh>
```


### Review process

The review process follows the GlusterFS review process.

1. Every new change will be developed in a new branch

2. Changes can be posted for review using the `git-review` tool

3. Reviews happen on GerritHub.

4. You iterate the change based on reviews and keep pushing patchsets.

5. The change finally gets merged and will be available from the Github repository

