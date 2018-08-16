package restclient

import (
	"github.com/gluster/glusterd2/pkg/api"
	bitrotapi "github.com/gluster/glusterd2/plugins/bitrot/api"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"
	gshdapi "github.com/gluster/glusterd2/plugins/glustershd/api"
)

// GlusterD2Client contains methods for GlusterD2 APIs
type GlusterD2Client interface {
	BitrotClient
	DeviceClient
	EventsClient
	GeorepClient
	VolumeClient
	PeerClient
	QuotaClient
	SnapshotClient
	SelfHealClient
	ClusterOptionClient
	Ping() error
	Version() (api.VersionResp, error)
}

// ClusterOptionClient defines methods for ClusterOption APIs
type ClusterOptionClient interface {
	ClusterOptionSet(req api.ClusterOptionReq) error
}

// BitrotClient defines methods for Bitrot APIs
type BitrotClient interface {
	BitrotEnable(volname string) error
	BitrotDisable(volname string) error
	BitrotScrubOndemand(volname string) error
	BitrotScrubStatus(volname string) (bitrotapi.ScrubStatus, error)
}

// DeviceClient defines methods for Device APIs
type DeviceClient interface {
	DeviceAdd(peerid string, device string) (deviceapi.AddDeviceResp, error)
}

// EventsClient methods for Events APIs
type EventsClient interface {
	WebhookAdd(url string, token string, secret string) error
	WebhookDelete(url string) error
	Webhooks() (eventsapi.WebhookList, error)
	ListEvents() ([]*api.Event, error)
	WebhookTest(url string, token string, secret string) error
}

// GeorepClient defines methods for Georeplication APIs
type GeorepClient interface {
	GeorepCreate(mastervolid string, slavevolid string, req georepapi.GeorepCreateReq) (georepapi.GeorepSession, error)
	GeorepStart(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, error)
	GeorepPause(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, error)
	GeorepResume(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, error)
	GeorepStop(mastervolid string, slavevolid string, force bool) (georepapi.GeorepSession, error)
	GeorepDelete(mastervolid string, slavevolid string, force bool) error
	GeorepStatus(mastervolid string, slavevolid string) (georepapi.GeorepSessionList, error)
	GeorepSSHKeysGenerate(volname string) ([]georepapi.GeorepSSHPublicKey, error)
	GeorepSSHKeys(volname string) ([]georepapi.GeorepSSHPublicKey, error)
	GeorepSSHKeysPush(volname string, sshkeys []georepapi.GeorepSSHPublicKey) error
	GeorepGet(mastervolid string, slavevolid string) ([]georepapi.GeorepOption, error)
	GeorepSet(mastervolid string, slavevolid string, keyvals map[string]string) error
	GeorepReset(mastervolid string, slavevolid string, keys []string) error
}

// SelfHealClient defines methods for SelfHeal APIs
type SelfHealClient interface {
	SelfHealInfo(params ...string) ([]gshdapi.BrickHealInfo, error)
	SelfHeal(volname string, healType string) error
}

// VolumeClient defines methods for Volume APIs
type VolumeClient interface {
	VolumeCreate(req api.VolCreateReq) (api.VolumeCreateResp, error)
	Volumes(volname string, filterParams ...map[string]string) (api.VolumeListResp, error)
	BricksStatus(volname string) (api.BricksStatusResp, error)
	VolumeStatus(volname string) (api.VolumeStatusResp, error)
	VolumeStart(volname string, force bool) error
	VolumeStop(volname string) error
	VolumeDelete(volname string) error
	VolumeSet(volname string, req api.VolOptionReq) error
	VolumeGet(volname string, optname string) (api.VolumeOptionsGetResp, error)
	VolumeExpand(volname string, req api.VolExpandReq) (api.VolumeExpandResp, error)
	VolumeStatedump(volname string, req api.VolStatedumpReq) error
	OptionGroupCreate(req api.OptionGroupReq) error
	OptionGroupList() (api.OptionGroupListResp, error)
	OptionGroupDelete(group string) error
	EditVolume(volname string, req api.VolEditReq) (api.VolumeEditResp, error)
	VolumeReset(volname string, req api.VolOptionResetReq) error
}

// PeerClient defines methods for Peer APIs
type PeerClient interface {
	PeerAdd(peerAddReq api.PeerAddReq) (api.PeerAddResp, error)
	PeerRemove(peerid string) error
	GetPeer(peerid string) (api.PeerGetResp, error)
	Peers(filterParams ...map[string]string) (api.PeerListResp, error)
}

// QuotaClient defines methods for Quota APIs
type QuotaClient interface {
	QuotaEnable(volname string) error
}

// SnapshotClient defines methods for Snapshot APIs
type SnapshotClient interface {
	SnapshotCreate(req api.SnapCreateReq) (api.SnapCreateResp, error)
	SnapshotActivate(req api.SnapActivateReq, snapname string) error
	SnapshotDeactivate(snapname string) error
	SnapshotList(volname string) (api.SnapListResp, error)
	SnapshotInfo(snapname string) (api.SnapGetResp, error)
	SnapshotDelete(snapname string) error
	SnapshotStatus(snapname string) (api.SnapStatusResp, error)
	SnapshotRestore(snapname string) (api.VolumeGetResp, error)
	SnapshotClone(snapname string, req api.SnapCloneReq) (api.VolumeCreateResp, error)
}

// This will ensure that Client always implements GlusterdD2Client interface
var _ GlusterD2Client = &Client{}
