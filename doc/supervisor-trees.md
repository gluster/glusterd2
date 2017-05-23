# Supervisor trees in GD2

GD2 uses supervisor trees to manage its internal services, which provides a clean and consistent pattern to implement these services.

GD2 uses the [`suture`][1] package to implement its supervisor tree, which was inspired by the Erlang/Elixir supervisor trees. More information on supervisor trees and the `suture` package can be found in [this blog post][2].

## Quick introduction

Supervisor trees involve 2 types of objects, supervisors and services.

### Services

Services are objects that provide or perform actual services. Services are managed by supervisors, and can be added to only one supervisor at a time. Services are started and managed by supervisors.

### Supervisors

Supervisors are a specialized services, which manage other services. Supervisors manage the lifecyle of the services that have been added to them. They take care of starting, restarting, stopping and keeping the services alive.

Supervisors themselves can be added as services to other supervisors. This way supervisor trees can be built to manage lots of different services in the application.

Only the top-most (root) supervisor in a supervisor tree is started manually. All other supervisors will be started by their parent supervisors.


## The GD2 supervisor tree

Currently, supervisor trees are to manage the network services provides by GD2 (rest, grpc, sunrpc). In the future, services like etcd, plugins etc. can also be managed using supervisor trees.

The structure of the GD2 supervisor tree currently is as shown below.

```

  gd2-main
  |
  +-->gd2-servers
      |
      +-->peerrpc
      |
      +-->gd2-muxserver
          |
          |-->muxlistener
          |
          |-->rest
          |
          +-->sunrpc

```
> in the graph above, all the leaves are services and everything else supervisors.

- `gd2-main` is the root supervisor. It is created and started in the main function, and manages all other services.
- `gd2-servers` is the supervisor which manages the servers started by GD2. It is implemented in the `servers` package. It manages the muxserver and the peerrpc server.
- `peerrpc` is the gRPC server used for internal communications.
- `gd2-muxserver` is a supervisor managing the muxed listener and the services listening on the muxed listener.
- `muxlistener` is the cmux multiplexed listener.
- `rest` is the GD2 rest server used for serving the management api.
- `sunrpc` is the GD2 sunrpc server used for serving mount and portmap requests for clients.

The supervisors and services are started top-down, as and when they get added to the tree and in parallel. Stopping also happens top-down, and in parallel. The root supervisor only returns once all its children have stoped.

[1]: https://github.com/thejerf/suture
[2]: http://www.jerf.org/iri/post/2930
