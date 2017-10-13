package georeplication

// GeorepCreateReq represents REST API request to create Geo-rep session
type GeorepCreateReq struct {
	MasterVol  string   `json:"mastervol"`
	SlaveUser  string   `json:"slaveuser"`
	SlaveHosts []string `json:"slavehosts"`
	SlaveVol   string   `json:"slavevol"`
}
