# Managing the glusterd Service

After installing GlusterFS and Glusterd2, you must start `glusterd2`
service. The glusterd2 service serves as the Gluster elastic volume
manager, overseeing glusterfs processes, and co-ordinating dynamic
volume operations, such as adding and removing volumes across multiple
storage servers non-disruptively.

This section describes how to start the glusterd2 service in the
following ways:

**Note**: You must start glusterd2 on all GlusterFS servers.

## Gluster user group

On each server nodes, create "gluster" user group if not exists
already using the following command.

    # groupadd gluster

Users from this group can run `glustercli` commands in GlusterFS
servers.

**Note**: Create "gluster" group before starting `glusterd2`

## Starting and Stopping glusterd2 Manually

This section describes how to start and stop glusterd2 manually

To start glusterd2 manually, enter the following command:

    # systemctl start glusterd2

To stop glusterd2 manually, enter the following command:

    # systemctl stop glusterd2

## Starting glusterd Automatically

This section describes how to configure the system to automatically
start the glusterd2 service every time the system boots.

    # systemctl enable glusterd2

## Status

Check the status of `glusterd2` using,

    # systemctl status glusterd2

To check the Cluster status, run the following command from one of the
GlusterFS server.

    # glustercli peer status
