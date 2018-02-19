
<!---
This file is generated using commands described below. DO NOT EDIT.

$ curl -o endpoints.json -s -X GET http://127.0.0.1:24007/endpoints
$ go build pkg/tools/generate-doc.go
$ ./generate-doc
-->

# REST API Endpoints Reference

Name | Methods | Path | Request | Response
--- | --- | --- | --- | ---
GetVersion | GET | /version | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [VersionResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VersionResp)
VolumeCreate | POST | /volumes | [VolCreateReq](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolCreateReq) | [VolumeCreateResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolumeCreateResp)
VolumeExpand | POST | /volumes/{volname}/expand | [VolExpandReq](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolExpandReq) | [VolumeExpandResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolumeExpandResp)
VolumeOptions | POST | /volumes/{volname}/options | [VolOptionReq](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolOptionReq) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
OptionGroupList | GET | /volumes/options-group | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [OptionGroupListResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#OptionGroupListResp)
OptionGroupCreate | POST | /volumes/options-group | [OptionGroupReq](https://godoc.org/github.com/gluster/glusterd2/pkg/api#OptionGroupReq) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
OptionGroupDelete | DELETE | /volumes/options-group/{groupname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
VolumeDelete | DELETE | /volumes/{volname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
VolumeInfo | GET | /volumes/{volname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [VolumeGetResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolumeGetResp)
VolumeBricksStatus | GET | /volumes/{volname}/bricks | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [BricksStatusResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#BricksStatusResp)
VolumeStatus | GET | /volumes/{volname}/status | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [VolumeStatusResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolumeStatusResp)
VolumeList | GET | /volumes | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [VolumeListResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#VolumeListResp)
VolumeStart | POST | /volumes/{volname}/start | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
VolumeStop | POST | /volumes/{volname}/stop | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
VolfilesGenerate | POST | /volfiles | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
VolfilesGet | GET | /volfiles | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GetPeer | GET | /peers/{peerid} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [PeerGetResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#PeerGetResp)
GetPeers | GET | /peers | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [PeerListResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#PeerListResp)
DeletePeer | DELETE | /peers/{peerid} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
AddPeer | POST | /peers | [PeerAddReq](https://godoc.org/github.com/gluster/glusterd2/pkg/api#PeerAddReq) | [PeerAddResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#PeerAddResp)
GeoreplicationCreate | POST | /geo-replication/{mastervolid}/{remotevolid} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationStart | POST | /geo-replication/{mastervolid}/{remotevolid}/start | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationStop | POST | /geo-replication/{mastervolid}/{remotevolid}/stop | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationDelete | DELETE | /geo-replication/{mastervolid}/{remotevolid} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationPause | POST | /geo-replication/{mastervolid}/{remotevolid}/pause | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationResume | POST | /geo-replication/{mastervolid}/{remotevolid}/resume | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationStatus | GET | /geo-replication/{mastervolid}/{remotevolid} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoReplicationConfigGet | GET | /geo-replication/{mastervolid}/{remotevolid}/config | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoReplicationConfigSet | POST | /geo-replication/{mastervolid}/{remotevolid}/config | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoReplicationConfigReset | DELETE | /geo-replication/{mastervolid}/{remotevolid}/config | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationStatusList | GET | /geo-replication | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationSshKeyGenerate | POST | /ssh-key/{volname}/generate | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationSshKeyPush | POST | /ssh-key/{volname}/push | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GeoreplicationSshKeyGet | GET | /ssh-key/{volname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
BitrotEnable | POST | /volumes/{volname}/bitrot/enable | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
BitrotDisable | POST | /volumes/{volname}/bitrot/disable | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
ScrubOndemand | POST | /volumes/{volname}/bitrot/scrubondemand | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
ScrubStatus | GET | /volumes/{volname}/bitrot/scrubstatus | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
QuotaEnable | POST | /quota/{volname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
QuotaDisable | DELETE | /quota/{volname} | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
QuotaList | GET | /quota/{volname}/limit | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
QuotaLimit | POST | /quota/{volname}/limit | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
QuotaRemove | DELETE | /quota/{volname}/limit | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
WebhookAdd | POST | /events/webhook | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
WebhookDelete | DELETE | /events/webhook | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
WebhookList | GET | /events/webhook | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GlustershEnable | POST | /volumes/{name}/heal/enable | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
GlustershDisable | POST | /volumes/{name}/heal/disable | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
Statedump | GET | /statedump | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#)
List Endpoints | GET | /endpoints | [](https://godoc.org/github.com/gluster/glusterd2/pkg/api#) | [ListEndpointsResp](https://godoc.org/github.com/gluster/glusterd2/pkg/api#ListEndpointsResp)
