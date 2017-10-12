package georeplication

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
)

// GeorepWorker represents Geo-replication Worker
type GeorepWorker struct {
	MasterNode                 string `json:"master_node"`
	MasterNodeID               string `json:"node_id"`
	MasterBrickPath            string `json:"master_brick_path"`
	MasterBrick                string `json:"master_brick"`
	Status                     string `json:"status"`
	LastSyncedTime             string `json:"last_synced_time"`
	LastSyncedTimeUTC          string `json:"last_synced_time_utc"`
	LastEntrySyncedTime        string `json:"last_synced_entry"`
	SlaveNode                  string `json:"slave_node"`
	ChangeDetection            string `json:"change_detection"`
	CheckpointTime             string `json:"checkpoint_time"`
	CheckpointTimeUTC          string `json:"checkpoint_time_utc"`
	CheckpointCompleted        bool   `json:"checkpoint_completed"`
	CheckpointCompletedTime    string `json:"checkpoint_completed_time"`
	CheckpointCompletedTimeUTC string `json:"checkpoint_completed_time_utc"`
	MetaOps                    string `json:"meta"`
	EntryOps                   string `json:"entry"`
	DataOps                    string `json:"data"`
	FailedOps                  string `json:"failures"`
	CrawlStatus                string `json:"crawl_status"`
}

// GeorepSession represents Geo-replication session
type GeorepSession struct {
	MasterID   uuid.UUID         `json:"master_volume_id"`
	SlaveID    uuid.UUID         `json:"slave_volume_id"`
	MasterVol  string            `json:"master_volume"`
	SlaveUser  string            `json:"slave_user"`
	SlaveHosts []string          `json:"slave_hosts"`
	SlaveVol   string            `json:"slave_volume"`
	Status     string            `json:"monitor_status"`
	Workers    []GeorepWorker    `json:"workers"`
	Options    map[string]string `json:"options"`
}

// NewGeorepSession creates new instance of GeorepSession
func NewGeorepSession(mastervolid uuid.UUID, slavevolid uuid.UUID, req GeorepCreateReq) *GeorepSession {
	slaveUser := req.SlaveUser
	if req.SlaveUser == "" {
		slaveUser = "root"
	}
	return &GeorepSession{
		MasterID:   mastervolid,
		SlaveID:    slavevolid,
		MasterVol:  req.MasterVol,
		SlaveVol:   req.SlaveVol,
		SlaveHosts: req.SlaveHosts,
		SlaveUser:  slaveUser,
		Status:     GeorepStatusCreated,
		Workers:    []GeorepWorker{},
		Options:    make(map[string]string),
	}
}
