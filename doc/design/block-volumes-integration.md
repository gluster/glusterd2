# Managing Gluster Block Volumes

Following new APIs will be introduced to support Gluster block Volumes.

## New Config and Environment variable options and its default values

Environment variables will get more priority than defined in config
file.

```
block-hosting-volume-size          = 500G
block-hosting-volume-type          = Replica
block-hosting-volume-replica-count = 3
auto-create-block-hosting-volumes  = false
```

Note: Environment variables will starts with `GD2_`, for example
`GD2_BLOCK_HOSTING_VOLUME_SIZE`

## Create a Block Volume

```
POST /v1/blockvolumes
```

gluster-block project provides CLI command to create, delete and list
the block volumes. Glusterd2 will provide wrapper around these CLI
commands to provide block volumes.

Request format:


```golang
type BlockVolumeCreateRequest struct {
    // HostingVolume name is optional
    HostingVolume string
    // Name represents block Volume name
    Name       string    `json:"name"`
    // Size represents Block Volume size in bytes
    Size       int       `json:"size"`
    Clusters   []string  `json:"clusters,omitempty"`
    HaCount    int       `json:"hacount,omitempty"`
    Auth       bool      `json:"auth,omitempty"
}
```

- If `HostingVolume` name is not empty, then run `gluster-block
  create` command to create block volume with requested size. If
  available size is less than requested size then ERROR. Set block
  related metadata and volume options if not exists.

- If `HostingVolume` is not specified, that means a Block hosting
  volume may needs to be created. List all available volumes and see
  if any volume is available with Metadata:`block-hosting=yes`.

- If No volumes are available with Metadata:`block-hosting=yes` or if
  no space available to create block
  volumes(Metadata:`block-hosting-available-size` is less than
  request size), then try to create a new block hosting Volume with
  generated name with default size and volume type configured. Also
  set required Volume options for Block hosting Volume.
  (Note: ERROR if `auto-create-block-hosting-volumes=false`)

- Set Metadata:`block-hosting-volume-auto-created=yes` if Block
  hosting volume is created.

- If no space available to create new block hosting Volume - ERROR

- Get the list the Gluster Volumes where Metadata:`block-hosting=yes`

- Look for already existing block Volume. Run `gluster-block info
  <block-hosting-vol>/<block-vol-name>` command for each block hosting
  volume. ERROR if Volume name already exists. (This may become
  expensive if number of Block hosting volumes are many. TBD)

- Pick a Block hosting Volume if
  Metadata:`block-hosting-available-size` is greater than request
  size.

- If no block hosting volume is available to create a block volume
  with requested size then ERROR.

- Run `gluster-block create` command to create the block volume with
  requested size.

- Update Metadata:`block-hosting-available-size` by subtracting the
  requested size.

## Listing the Block Volumes

```
GET /v1/blockvolumes
GET /v1/blockvolumes/:name
```

- Get the list of Block hosting Volumes by listing the Gluster volumes
  with Metadata:`block-hosting=yes`
- Run `gluster-block list` CLI command for each block hosting volumes
  and aggregate the response.

## Deleting the Block Volume

```
DELETE /v1/blockvolumes/:name
```

- Get list of Block hosting volumes by listing the Gluster volumes
  with Metadata:`block-hosting=yes`
- Run `gluster-block delete` command for each block hosting
  volume.(Break after it finds a requested block volume)
- Update Metadata:`block-hosting-available-size` by adding the
  deleted volume size.
- If number of block volumes in the block hosting volume is zero then
  stop and delete the Block hosting Volume. (Note: Delete block
  hosting volume only if Metadata:`block-hosting-volume-auto-created=yes`)


## References

- Gluster Block project page https://github.com/gluster/gluster-block
- Gluster block Heketi Integration
  https://github.com/gluster/gluster-kubernetes/blob/master/docs/design/gluster-block-provisioning.md
