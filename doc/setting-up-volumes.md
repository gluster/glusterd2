# Setting up GlusterFS Volumes

The commands that differ with GD2 are mentioned in this doc. For info about volume types and so on you can refer [here](https://docs.gluster.org/en/latest/Administrator%20Guide/Setting%20Up%20Volumes/)

## Creating New Volumes:

### Creating Distributed Volumes

    `# glustercli volume create --name <VOLNAME> <UUID1>:<brick1> .. <UUIDn>:<brickm> `

    where n is the number of servers and m is the number of bricks. n and m can be same or m can be more than n.

    For example, a four node distributed volume:

        # glustercli volume create --name testvol server1:/export/brick1/data server2:/export/brick2/data server3:/export/brick3/data server4:/export/brick4/data
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

### Creating Replicated Volumes

    `# glustercli volume create --name <VOLNAME> --replica <count> <UUID1>:<brick1> .. <UUIDn>:<brickm>`

    where n is the server count and m is the number of bricks.

    For example, to create a replicated volume with two storage servers:

        # glustercli volume create testvol server1:/exp1 server2:/exp2 --replica 2
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

    > **Note**:

    > - GlusterD2 creates a replicate volume if more than one brick of a replica set is present on the same peer. For eg. a four node replicated volume where more than one brick of a replica set is present on the same peer.
    >

    >         # glustercli volume create --name <VOLNAME> --replica 4 server1:/brick1 server1:/brick2 server2:/brick2 server3:/brick3
    >           <VOLNAME> Volume created successfully
    >           Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

### Arbiter configuration for replica volumes

    '# glustercli volume create <VOLNAME> --replica 2 --arbiter 1 <UUID1>:<brick1> <UUID2>:<brick2> <UUID3>:<brick3>'

>**Note:**
>
>       1) It is mentioned as replica 2 and not replica 3 even though there are 3 replicas (arbiter included).
>       2) The arbiter configuration for replica 3 can be used to create distributed-replicate volumes as well.

## Creating Distributed Replicated Volumes

    `# glustercli volume create --name <VOLNAME> <UUID1>:<brick1> .. <UUIDn>:<brickm> --replica <count> `

    where n is the number of servers and m is the number of bricks.

    For example, a four node distributed (replicated) volume with a
    two-way mirror:

        # glustercli volume create --name testvol server1:/export/brick1/data server2:/export/brick2/data server1:/export/brick3/data server2:/export/brick4/data --replica 2
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

    For example, to create a six node distributed (replicated) volume
    with a two-way mirror:

        # glustercli volume create testvol server1:/exp1 server2:/exp2 server3:/exp3 server4:/exp4 server5:/exp5 server6:/exp6 --replica 2
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

    > **Note**:

    > - GlusterD2 creates a distribute replicate volume if more than one brick of a replica set is present on the same peer. For eg. for a four node distribute (replicated) volume where more than one brick of a replica set is present on the same peer.
    >

    >         # glustercli volume create --name <volname> --replica 2 server1:/brick1 server1:/brick2 server2:/brick3 server2:/brick4
    >           <VOLNAME> Volume created successfully
    >           Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99


## Creating Dispersed Volumes

    `# glustercli volume create --name <VOLNAME> --disperse <COUNT> <UUID1>:<brick1> .. <UUIDn>:<brickm>`

    For example, a four node dispersed volume:

        # glustercli volume create --name testvol --dispersed 4 server{1..4}:/export/brick/data
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

    For example, to create a six node dispersed volume:

        # glustercli volume create testvol --disperse 6 server{1..6}:/export/brick/data
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99

        The redundancy count is automatically set as 2 here.

## Creating Distributed Dispersed Volumes

    `# glustercli volume create --name <VOLNAME> --disperse <COUNT> <UUID1>:<brick1> .. <UUIDn>:<brickm>`

    For example, to create a six node dispersed volume:

        # glustercli volume create testvol --disperse 3 server1:/export/brick/data{1..6}
        testvol Volume created successfully
        Volume ID:  15c1611d-aae6-44f0-ae8d-fa04f31f5c99


## Starting Volumes

You must start your volumes before you try to mount them.

**To start a volume**

-   Start a volume:

    `# glustercli volume start <VOLNAME>`

    For example, to start test-volume:

        # glustercli volume start testvol
        Volume testvol started successfully
