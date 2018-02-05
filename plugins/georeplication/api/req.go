package georeplication

// GeorepSlaveHostReq represents Slave host ID and IP/Hostname
type GeorepSlaveHostReq struct {
	NodeID   string `json:"nodeid"`
	Hostname string `json:"host"`
}

// GeorepCreateReq represents REST API request to create Geo-rep session
type GeorepCreateReq struct {
	MasterVol  string               `json:"mastervol"`
	SlaveUser  string               `json:"slaveuser"`
	SlaveHosts []GeorepSlaveHostReq `json:"slavehosts"`
	SlaveVol   string               `json:"slavevol"`
}
