# Managing Trusted Storage Pools


### Overview

A trusted storage pool(TSP) is a trusted network of storage servers(peers). More about TSP [here](https://docs.gluster.org/en/latest/Administrator%20Guide/Storage%20Pools/)

The respective commands for glusterd2 can be found below.


-  [Adding Servers](#adding-servers)
-  [Listing Servers](#listing-servers)
-  [Viewing Peer Status](#peer-status)
-  [Removing Servers](#removing-servers)


<a name="adding-servers"></a>
### Adding Servers

To add a server to a TSP, do add peer from a server already in the pool.

        # glustercli peer add <server>

For example, to add a new server(server2) to the cluster described above, probe it from one of the other servers:

        server1# glustercli peer add server2
        Peer add successful
        +--------------------------------------+---------+-----------------------+-----------------------+
        |                  ID                  |  NAME   |   CLIENT ADDRESSES    |    PEER ADDRESSES     |
        +--------------------------------------+---------+-----------------------+-----------------------+
        | fd0aaa07-9e5f-4265-b778-e49514874ca2 | server2 | 127.0.0.1:24007       | server2:24008         |
        |                                      |         | 192.168.122.193:24007 | 192.168.122.193:24008 |
        +--------------------------------------+---------+-----------------------+-----------------------+


Verify the peer status from the first server (server1):

        server1# glustercli peer status
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        |                  ID                  |  NAME   |   CLIENT ADDRESSES    |    PEER ADDRESSES     | ONLINE |  PID  |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        | d82734dc-57c0-44ef-a682-8b59c43d0cef | server1 | 127.0.0.1:24007       | 192.168.122.18:24008  | yes    |  1269 |
        |                                      |         | 192.168.122.18:24007  |                       |        |       |
        | fd0aaa07-9e5f-4265-b778-e49514874ca2 | server2 | 127.0.0.1:24007       | 192.168.122.193:24008 | yes    | 18657 |
        |                                      |         | 192.168.122.193:24007 |                       |        |       |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+


<a name="listing-servers"></a>
### Listing Servers

To list all nodes in the TSP:

        server1# glustercli peer list
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        |                  ID                  |  NAME   |   CLIENT ADDRESSES    |    PEER ADDRESSES     | ONLINE |  PID  |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        | d82734dc-57c0-44ef-a682-8b59c43d0cef | server1 | 127.0.0.1:24007       | 192.168.122.18:24008  | yes    |  1269 |
        |                                      |         | 192.168.122.18:24007  |                       |        |       |
        | fd0aaa07-9e5f-4265-b778-e49514874ca2 | server2 | 127.0.0.1:24007       | 192.168.122.193:24008 | yes    | 18657 |
        |                                      |         | 192.168.122.193:24007 |                       |        |       |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+


<a name="peer-status"></a>
### Viewing Peer Status

To view the status of the peers in the TSP:

        server1# glustercli peer status
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        |                  ID                  |  NAME   |   CLIENT ADDRESSES    |    PEER ADDRESSES     | ONLINE |  PID  |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+
        | d82734dc-57c0-44ef-a682-8b59c43d0cef | server1 | 127.0.0.1:24007       | 192.168.122.18:24008  | yes    |  1269 |
        |                                      |         | 192.168.122.18:24007  |                       |        |       |
        | fd0aaa07-9e5f-4265-b778-e49514874ca2 | server2 | 127.0.0.1:24007       | 192.168.122.193:24008 | yes    | 18657 |
        |                                      |         | 192.168.122.193:24007 |                       |        |       |
        +--------------------------------------+---------+-----------------------+-----------------------+--------+-------+


<a name="removing-servers"></a>
### Removing Servers

To remove a server from the TSP, run the following command from another server in the pool:

        # gluster peer remove <peer-ID>

For example, to remove server4 from the trusted storage pool:

        server1# glustercli peer remove fd0aaa07-9e5f-4265-b778-e49514874ca2
        Peer remove success

***Note:*** For now remove peer works only with peerid which you can get from peer status.

Verify the peer status:

        server1# glustercli peer status
        +--------------------------------------+---------+----------------------+----------------------+--------+------+
        |                  ID                  |  NAME   |   CLIENT ADDRESSES   |    PEER ADDRESSES    | ONLINE | PID  |
        +--------------------------------------+---------+----------------------+----------------------+--------+------+
        | d82734dc-57c0-44ef-a682-8b59c43d0cef | server1 | 127.0.0.1:24007      | 192.168.122.18:24008 | yes    | 1269 |
        |                                      |         | 192.168.122.18:24007 |                      |        |      |
        +--------------------------------------+---------+----------------------+----------------------+--------+------+

