package api

// BlockVolumeInfo represents block volume info
type BlockVolumeInfo struct {
	// HostingVolume name is optional
	HostingVolume string `json:"hostingvolume"`
	// Name represents block Volume name
	Name string `json:"name"`
	// Size represents Block Volume size in bytes
	Size    uint64 `json:"size,omitempty"`
	HaCount int    `json:"hacount,omitempty"`
}

// HostVolumeInfo represents block volume info
type HostVolumeInfo struct {
	// HostVolReplicaCnt represents the replica count of the block hosting volume
	HostVolReplicaCnt int `json:"hostvolreplicacnt,omitempty"`
	// HostVolThinArbPath represents the thin arbiter path
	HostVolThinArbPath string `json:"hostvolthinarbpath,omitempty"`
	// HostVolShardSize represents the shard size of the block hosting volume
	HostVolShardSize uint64 `json:"hostvolshardsize,omitempty"`
	// Size represents Block Hosting Volume size in bytes
	HostVolSize uint64 `json:"hostvolsize,omitempty"`
}

// BlockVolumeCreateRequest represents req Body for Block vol create req
type BlockVolumeCreateRequest struct {
	*BlockVolumeInfo
	HostVolumeInfo
	BlockType string   `json:"blocktype,omitempty"`
	Clusters  []string `json:"clusters,omitempty"`
	Auth      bool     `json:"auth,omitempty"`
}

// BlockVolumeCreateResp represents resp body for a Block Vol Create req
type BlockVolumeCreateResp struct {
	*BlockVolumeInfo
	Hosts    []string `json:"hosts"`
	Iqn      string   `json:"iqn"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}

// BlockVolumeListResp represents resp body for a Block List req
type BlockVolumeListResp []BlockVolumeInfo

// BlockVolumeGetResp represents a resp body for Block Vol Get req
type BlockVolumeGetResp struct {
	*BlockVolumeInfo
	Hosts    []string `json:"hosts"`
	GBID     string   `json:"gbid"`
	Password string   `json:"password,omitempty"`
}
