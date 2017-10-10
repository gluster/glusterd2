package georeplication

import (
	"github.com/pborman/uuid"
)

const (
	georepStatusCreated = "Created"
	georepStatusStarted = "Started"
	georepStatusStopped = "Stopped"
	georepStatusPaused  = "Paused"
)

// Worker represents Geo-replication Worker
type Worker struct {
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

// Session represents Geo-replication session
type Session struct {
	MasterID   uuid.UUID         `json:"master_volume_id"`
	SlaveID    uuid.UUID         `json:"slave_volume_id"`
	MasterVol  string            `json:"master_volume"`
	SlaveUser  string            `json:"slave_user"`
	SlaveHosts []string          `json:"slave_hosts"`
	SlaveVol   string            `json:"slave_volume"`
	Status     string            `json:"monitor_status"`
	Workers    []Worker          `json:"workers"`
	Options    map[string]string `json:"options"`
}

// GeorepCreateRequest represents REST API request to create Geo-rep session
type GeorepCreateRequest struct {
	MasterVol  string   `json:"mastervol"`
	SlaveUser  string   `json:"slaveuser"`
	SlaveHosts []string `json:"slavehosts"`
	SlaveVol   string   `json:"slavevol"`
}
