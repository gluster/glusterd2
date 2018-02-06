package georeplication

// GeorepRemoteHostReq represents Remote host ID and IP/Hostname
type GeorepRemoteHostReq struct {
	NodeID   string `json:"nodeid"`
	Hostname string `json:"host"`
}

// GeorepCreateReq represents REST API request to create Geo-rep session
type GeorepCreateReq struct {
	MasterVol   string                `json:"mastervol"`
	RemoteUser  string                `json:"remoteuser"`
	RemoteHosts []GeorepRemoteHostReq `json:"remotehosts"`
	RemoteVol   string                `json:"remotevol"`
	Force       bool                  `json:"force"`
}

// GeorepCommandsReq represents extra arguments to Geo-rep APIs
type GeorepCommandsReq struct {
	Force bool `json:"force"`
}
