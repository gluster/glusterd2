package api

import (
	"github.com/pborman/uuid"
)

const (
	// GeorepStatusCreated represents Created State
	GeorepStatusCreated = "Created"

	// GeorepStatusStarted represents Started State
	GeorepStatusStarted = "Started"

	// GeorepStatusStopped represents Stopped State
	GeorepStatusStopped = "Stopped"

	// GeorepStatusPaused represents Paused State
	GeorepStatusPaused = "Paused"

	// GeorepStatusInitializing represents worker initializing state
	GeorepStatusInitializing = "Initializing.."

	// GeorepStatusActive represents Active worker
	GeorepStatusActive = "Active"

	// GeorepStatusPassive represents Passive worker
	GeorepStatusPassive = "Passive"

	// GeorepStatusUnknown represents worker's state Unknown(If Glusterd of that node is not reachable)
	GeorepStatusUnknown = "Unknown"

	// GeorepStatusFaulty represents Faulty worker
	GeorepStatusFaulty = "Faulty"
)

// GeorepRemoteHost represents Remote host UUID and Hostname
type GeorepRemoteHost struct {
	PeerID   uuid.UUID `json:"peerid"`
	Hostname string    `json:"host"`
}

// GeorepWorker represents Geo-replication Worker
type GeorepWorker struct {
	MasterPeerHostname         string `json:"master_peer_hostname"`
	MasterPeerID               string `json:"peer_id"`
	MasterBrickPath            string `json:"master_brick_path"`
	MasterBrick                string `json:"master_brick"`
	Status                     string `json:"worker_status"`
	LastSyncedTime             string `json:"last_synced"`
	LastSyncedTimeUTC          string `json:"last_synced_utc"`
	LastEntrySyncedTime        string `json:"last_synced_entry"`
	RemotePeerHostname         string `json:"remote_peer_hostname"`
	CheckpointTime             string `json:"checkpoint_time"`
	CheckpointTimeUTC          string `json:"checkpoint_time_utc"`
	CheckpointCompleted        string `json:"checkpoint_completed"`
	CheckpointCompletedTime    string `json:"checkpoint_completion_time"`
	CheckpointCompletedTimeUTC string `json:"checkpoint_completion_time_utc"`
	MetaOps                    string `json:"meta"`
	EntryOps                   string `json:"entry"`
	DataOps                    string `json:"data"`
	FailedOps                  string `json:"failures"`
	CrawlStatus                string `json:"crawl_status"`
}

// GeorepSSHPublicKey represents one nodes SSH Public key
type GeorepSSHPublicKey struct {
	PeerID    uuid.UUID `json:"peerid"`
	GsyncdKey string    `json:"gsyncd"`
	TarKey    string    `json:"tar"`
}

// GeorepSession represents Geo-replication session
type GeorepSession struct {
	MasterID    uuid.UUID          `json:"master_volume_id"`
	RemoteID    uuid.UUID          `json:"remote_volume_id"`
	MasterVol   string             `json:"master_volume"`
	RemoteUser  string             `json:"remote_user"`
	RemoteHosts []GeorepRemoteHost `json:"remote_hosts"`
	RemoteVol   string             `json:"remote_volume"`
	Status      string             `json:"monitor_status"`
	Workers     []GeorepWorker     `json:"workers"`
	Options     map[string]string  `json:"options"`
}

// GeorepOption represents Config details
type GeorepOption struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	DefaultValue string `json:"default_value"`
	Configurable bool   `json:"configurable"`
	Modified     bool   `json:"modified"`
}
