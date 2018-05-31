package api

// VolCreateReq represents a Volume Create Request
type VolCreateReq struct {
	Name                    string   `json:"name"`
	Transport               string   `json:"transport,omitempty"`
	Force                   bool     `json:"force,omitempty"`
	Size                    uint64   `json:"size"`
	DistributeCount         int      `json:"distribute"`
	ReplicaCount            int      `json:"replica"`
	ArbiterCount            int      `json:"arbiter"`
	DisperseCount           int      `json:"disperse"`
	DisperseRedundancyCount int      `json:"disperse-redundancy"`
	DisperseDataCount       int      `json:"disperse-data"`
	SnapshotEnabled         bool     `json:"snapshot"`
	SnapshotReserveFactor   float64  `json:"snapshot-reserve-factor"`
	LimitPeers              []string `json:"limit-peers"`
	LimitZones              []string `json:"limit-zones"`
	ExcludePeers            []string `json:"exclude-peers"`
	ExcludeZones            []string `json:"exclude-zones"`
	SubvolZonesOverlap      bool     `json:"subvolume-zones-overlap"`
}

// Brick represents Brick Request
type Brick struct {
	Type           string `json:"type"`
	PeerID         string `json:"peerid"`
	Path           string `json:"path"`
	TpMetadataSize uint64 `json:"metadata-size"`
	TpSize         uint64 `json:"thinkpool-size"`
	VgName         string `json:"vg-name"`
	TpName         string `json:"thinpool-name"`
	LvName         string `json:"logical-volume"`
	Size           uint64 `json:"size"`
	VgID           string `json:"vg-id"`
}

// Subvol represents Sub volume Request
type Subvol struct {
	Type                    string   `json:"type"`
	Bricks                  []Brick  `json:"bricks"`
	Subvols                 []Subvol `json:"subvols"`
	ReplicaCount            int      `json:"replica"`
	ArbiterCount            int      `json:"arbiter"`
	DisperseCount           int      `json:"disperse-count,omitempty"`
	DisperseDataCount       int      `json:"disperse-data,omitempty"`
	DisperseRedundancyCount int      `json:"disperse-redundancy,omitempty"`
}

// Volume represents a Volume Create Request
type Volume struct {
	Name                    string   `json:"name"`
	Transport               string   `json:"transport,omitempty"`
	Force                   bool     `json:"force,omitempty"`
	Size                    uint64   `json:"size"`
	DistributeCount         int      `json:"distribute"`
	ReplicaCount            int      `json:"replica"`
	ArbiterCount            int      `json:"arbiter"`
	DisperseCount           int      `json:"disperse"`
	DisperseDataCount       int      `json:"disperse-data"`
	DisperseRedundancyCount int      `json:"disperse-redundancy"`
	SnapshotEnabled         bool     `json:"snapshot"`
	SnapshotReserveFactor   float64  `json:"snapshot-reserve-factor"`
	LimitPeers              []string `json:"limit-peers"`
	LimitZones              []string `json:"limit-zones"`
	ExcludePeers            []string `json:"exclude-peers"`
	ExcludeZones            []string `json:"exclude-zones"`
	SubvolZonesOverlap      bool     `json:"subvolume-zones-overlap"`
	Subvols                 []Subvol `json:"subvols"`
	SubvolType              string   `json:"subvolume-type"`
}
