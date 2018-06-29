# Gluster Management APIs
Gluster Management APIs
## Get Peers information
**URL**

`/peers`

**HTTP Method**

`get`

**Parameters**

None

**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X GET http://localhost:24007/v1/peers
```
**Sample Response**

Status Code: 200

```json
[
    {
        "id": "3203f226-b37f-42da-83fe-0fe8d043d995",
        "name": "node1.example.com",
        "peer-addresses": [
            "node1.example.com"
        ],
        "client-addresses": [
            "127.0.0.1:24007",
            "192.168.122.12:24007"
        ],
        "online": true,
        "metadata": {
            "_zone": "3203f226-b37f-42da-83fe-0fe8d043d995"
        }
    },
    {
        "id": "a185f11a-630e-4776-995c-ff2cb84dfa91",
        "name": "node2.example.com",
        "peer-addresses": [
            "node2.example.com"
        ],
        "client-addresses": [
            "127.0.0.1:24007",
            "192.168.122.14:24007"
        ],
        "online": true,
        "metadata": {
            "_zone": "a185f11a-630e-4776-995c-ff2cb84dfa91"
        }
    },
    {
        "id": "ef27499e-9793-434d-b922-116ff4a315c7",
        "name": "node3.example.com",
        "peer-addresses": [
            "node3.example.com"
        ],
        "client-addresses": [
            "127.0.0.1:24007",
            "192.168.122.16:24007"
        ],
        "online": true,
        "metadata": {
            "_zone": "ef27499e-9793-434d-b922-116ff4a315c7"
        }
    }
]
```

## Add Peer
**URL**

`/peers`

**HTTP Method**

`post`

**Parameters**

None

**Request Body**(application/json)



| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| addresses | *Required*.  | array |
| zone | *Optional*.  | string |
| metadata | *Optional*.  | object |


**Sample Request**  :



```
curl -X POST http://localhost:24007/v1/peers -d '{"addresses": ["node1.example.com"]}'
```


**Sample Response**

Status Code: 201

```json
{
    "id": "3203f226-b37f-42da-83fe-0fe8d043d995",
    "name": "node1.example.com",
    "peer-addresses": [
        "node1.example.com"
    ],
    "client-addresses": [
        "127.0.0.1:24007",
        "192.168.122.12:24007"
    ],
    "online": true,
    "metadata": {
        "_zone": "3203f226-b37f-42da-83fe-0fe8d043d995"
    }
}
```

## Get a Peer information
**URL**

`/peers/{peerid}`

**HTTP Method**

`get`

**Parameters**


| Parameter | Description | Data Type |
| ---------- | ----------- | --------- |
| peerid | *Required*. Peer ID | string |


**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X GET http://localhost:24007/v1/peers/4f196836-0d9d-475a-aae2-642bb0eac685
```
**Sample Response**

Status Code: 200

```json
{
    "id": "3203f226-b37f-42da-83fe-0fe8d043d995",
    "name": "node1.example.com",
    "peer-addresses": [
        "node1.example.com"
    ],
    "client-addresses": [
        "127.0.0.1:24007",
        "192.168.122.12:24007"
    ],
    "online": true,
    "metadata": {
        "_zone": "3203f226-b37f-42da-83fe-0fe8d043d995"
    }
}
```

## Edit Peer
**URL**

`/peers/{peerid}`

**HTTP Method**

`post`

**Parameters**


| Parameter | Description | Data Type |
| ---------- | ----------- | --------- |
| peerid | *Required*. Peer ID | string |


**Request Body**(application/json)



| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| zone | *Optional*.  | string |
| metadata | *Optional*.  | object |


**Sample Request**  :



```
curl -X POST http://localhost:24007/v1/peers/4f196836-0d9d-475a-aae2-642bb0eac685 -d '{"metadata": [{"added_date": "2018-07-24"}]}'
```


**Sample Response**

Status Code: 200

```json
{
    "id": "3203f226-b37f-42da-83fe-0fe8d043d995",
    "name": "node1.example.com",
    "peer-addresses": [
        "node1.example.com"
    ],
    "client-addresses": [
        "127.0.0.1:24007",
        "192.168.122.12:24007"
    ],
    "online": true,
    "metadata": {
        "_zone": "3203f226-b37f-42da-83fe-0fe8d043d995"
    }
}
```

## Delete Peer
**URL**

`/peers/{peerid}`

**HTTP Method**

`del`

**Parameters**


| Parameter | Description | Data Type |
| ---------- | ----------- | --------- |
| peerid | *Required*. Peer ID | string |


**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X DEL http://localhost:24007/v1/peers/4f196836-0d9d-475a-aae2-642bb0eac685
```
**Sample Response**

Status Code: 204


## Get Volumes information
**URL**

`/volumes`

**HTTP Method**

`get`

**Parameters**

None

**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X GET http://localhost:24007/v1/volumes
```
**Sample Response**

Status Code: 200

```json
{
    "name": "gv1",
    "snap-list": [],
    "replica-count": 3,
    "arbiter-count": 0,
    "id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
    "state": "Started",
    "distribute-count": 1,
    "type": "Replicate",
    "options": {},
    "transport": "tcp",
    "metadata": {},
    "subvols": [
        {
            "name": "subvol-0",
            "type": "Replicate",
            "replica-count": 3,
            "arbiter-count": 0,
            "bricks": [
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node1.example.com",
                    "peer-id": "9bf0a19f-0680-438c-a213-ba16252c31da",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick1/brick",
                    "type": "brick",
                    "id": "b5f243c1-5705-4f19-acf0-fea36570b706"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node2.example.com",
                    "peer-id": "ea8a2787-4166-43ec-b5b3-e17fad649bae",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick2/brick",
                    "type": "brick",
                    "id": "e31ec08d-7855-40ad-a071-160e7aede43e"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node3.example.com",
                    "peer-id": "fd3d8424-7bea-4a75-a2a8-14f4bd450568",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick3/brick",
                    "type": "brick",
                    "id": "58cd6364-4245-414e-b0de-b27e380a73c9"
                }
            ]
        }
    ]
}
```

## Volume Create
**URL**

`/volumes`

**HTTP Method**

`post`

**Parameters**

None

**Request Body**(application/json)

Create Volume with bricks auto provisioned

| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| size | *Required*. Volume Size | int |
| name | *Optional*. Volume Name | string |
| transport | *Optional*. Transport Type | string |
| force | *Optional*. Force | bool |
| options | *Optional*. Options to be configured | object |
| advanced | *Optional*. Allow setting advanced options | bool |
| experimental | *Optional*. Allow setting experimental options | bool |
| deprecated | *Optional*. Allow setting deprecated options | bool |
| metadata | *Optional*. Set Volume Metadata | object |
| flags | *Optional*. Set Flags | object |
| distribute | *Optional*. Distribute count | int |
| replica | *Optional*. Replica Count | int |
| arbiter | *Optional*. Arbiter Count | int |
| disperse | *Optional*. Disperse count | int |
| disperse-redundancy | *Optional*. Disperse Redundancy count | int |
| disperse-data | *Optional*. Disperse Data count | int |
| snapshot | *Optional*. Enable Snapshot for the Volume | bool |
| snapshot-reserve-factor | *Optional*. Snapshot reserve factor | float |
| limit-peers | *Optional*. Create Volume only from these peers | array |
| limit-zones | *Optional*. Create Volume only from these zones | array |
| exclude-peers | *Optional*. Do not create Volume from these peers | array |
| exclude-zones | *Optional*. Do not create Volume from these zones | array |
| subvolume-zones-overlap | *Optional*. Bricks of different subvolume can be created on same device/peer/zone | bool |


**OR**
Create Volume with bricks manually provisioned

| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| name | *Optional*. Volume Name | string |
| transport | *Optional*. Transport Type | string |
| force | *Optional*. Force | bool |
| subvols | *Optional*. List of sub volumes | array |
| options | *Optional*. Options to be configured | object |
| advanced | *Optional*. Allow setting advanced options | bool |
| experimental | *Optional*. Allow setting experimental options | bool |
| deprecated | *Optional*. Allow setting deprecated options | bool |
| metadata | *Optional*. Set Volume Metadata | object |
| flags | *Optional*. Set Flags | object |

subvols

| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| type | *Required*. Sub volume Type | string |
| bricks | *Required*. List of Bricks in the sub volume | array |
| replica | *Optional*. Replica count | int |
| arbiter | *Optional*. Arbiter count | int |
| disperse-count | *Optional*. Disperse count | int |
| disperse-data | *Optional*. Disperse data count | int |
| disperse-redundancy | *Optional*. Disperse redundancy count | int |

bricks

| Field | Description | Data Type |
| ---------- | ----------- | --------- |
| type | *Optional*. Brick type | string |
| peerid | *Optional*. Peer ID | string |
| path | *Optional*. Brick Path | string |


**Sample Request**  :

Create Volume with bricks auto provisioned

```
curl -X POST http://localhost:24007/v1/volumes -d '{"name": "gv1", "size": 1000, "replica": 3}'
```

Create Volume with bricks manually provisioned

```
curl -X POST http://localhost:24007/v1/volumes -d '{"name": "gv1", "subvols": [{"type": "replicate", "replica": 3, "bricks": [{"peerid": "0c5bc279-397a-4535-be32-301c16dbbc69", "path": "/exports/bricks/gv1/brick1/brick"}, {"peerid": "7aafd270-b9b2-40b2-ba0e-7289f7d025c0", "path": "/exports/bricks/gv1/brick2/brick"}, {"peerid": "57470e13-2f9c-4404-9179-fb2ba38cc1d8", "path": "/exports/bricks/gv1/brick3/brick"}]}]}'
```


**Sample Response**

Status Code: 201

```json
{
    "name": "gv1",
    "snap-list": [],
    "replica-count": 3,
    "arbiter-count": 0,
    "id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
    "state": "Created",
    "distribute-count": 1,
    "type": "Replicate",
    "options": {},
    "transport": "tcp",
    "metadata": {},
    "subvols": [
        {
            "name": "subvol-0",
            "type": "Replicate",
            "replica-count": 3,
            "arbiter-count": 0,
            "bricks": [
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node1.example.com",
                    "peer-id": "9bf0a19f-0680-438c-a213-ba16252c31da",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick1/brick",
                    "type": "brick",
                    "id": "b5f243c1-5705-4f19-acf0-fea36570b706"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node2.example.com",
                    "peer-id": "ea8a2787-4166-43ec-b5b3-e17fad649bae",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick2/brick",
                    "type": "brick",
                    "id": "e31ec08d-7855-40ad-a071-160e7aede43e"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node3.example.com",
                    "peer-id": "fd3d8424-7bea-4a75-a2a8-14f4bd450568",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick3/brick",
                    "type": "brick",
                    "id": "58cd6364-4245-414e-b0de-b27e380a73c9"
                }
            ]
        }
    ]
}
```

## Start Volume
**URL**

`/volumes/{volname}/start`

**HTTP Method**

`post`

**Parameters**


| Parameter | Description | Data Type |
| ---------- | ----------- | --------- |
| volname | *Required*. Volume Name | string |


**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X POST http://localhost:24007/v1/volumes/gv1/start
```
**Sample Response**

Status Code: 200

```json
{
    "name": "gv1",
    "snap-list": [],
    "replica-count": 3,
    "arbiter-count": 0,
    "id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
    "state": "Started",
    "distribute-count": 1,
    "type": "Replicate",
    "options": {},
    "transport": "tcp",
    "metadata": {},
    "subvols": [
        {
            "name": "subvol-0",
            "type": "Replicate",
            "replica-count": 3,
            "arbiter-count": 0,
            "bricks": [
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node1.example.com",
                    "peer-id": "9bf0a19f-0680-438c-a213-ba16252c31da",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick1/brick",
                    "type": "brick",
                    "id": "b5f243c1-5705-4f19-acf0-fea36570b706"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node2.example.com",
                    "peer-id": "ea8a2787-4166-43ec-b5b3-e17fad649bae",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick2/brick",
                    "type": "brick",
                    "id": "e31ec08d-7855-40ad-a071-160e7aede43e"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node3.example.com",
                    "peer-id": "fd3d8424-7bea-4a75-a2a8-14f4bd450568",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick3/brick",
                    "type": "brick",
                    "id": "58cd6364-4245-414e-b0de-b27e380a73c9"
                }
            ]
        }
    ]
}
```

## Stop Volume
**URL**

`/volumes/{volname}/stop`

**HTTP Method**

`post`

**Parameters**


| Parameter | Description | Data Type |
| ---------- | ----------- | --------- |
| volname | *Required*. Volume Name | string |


**Request Body**(application/json)

None

**Sample Request**  :

```
curl -X POST http://localhost:24007/v1/volumes/gv1/stop
```
**Sample Response**

Status Code: 200

```json
{
    "name": "gv1",
    "snap-list": [],
    "replica-count": 3,
    "arbiter-count": 0,
    "id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
    "state": "Stopped",
    "distribute-count": 1,
    "type": "Replicate",
    "options": {},
    "transport": "tcp",
    "metadata": {},
    "subvols": [
        {
            "name": "subvol-0",
            "type": "Replicate",
            "replica-count": 3,
            "arbiter-count": 0,
            "bricks": [
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node1.example.com",
                    "peer-id": "9bf0a19f-0680-438c-a213-ba16252c31da",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick1/brick",
                    "type": "brick",
                    "id": "b5f243c1-5705-4f19-acf0-fea36570b706"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node2.example.com",
                    "peer-id": "ea8a2787-4166-43ec-b5b3-e17fad649bae",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick2/brick",
                    "type": "brick",
                    "id": "e31ec08d-7855-40ad-a071-160e7aede43e"
                },
                {
                    "volume-id": "95dd8a65-fc4b-447e-ba5b-8a541df319f2",
                    "host": "node3.example.com",
                    "peer-id": "fd3d8424-7bea-4a75-a2a8-14f4bd450568",
                    "volume-name": "gv1",
                    "path": "/exports/bricks/gv1/brick3/brick",
                    "type": "brick",
                    "id": "58cd6364-4245-414e-b0de-b27e380a73c9"
                }
            ]
        }
    ]
}
```

