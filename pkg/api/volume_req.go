package api

// BrickReq represents Brick Request
type BrickReq struct {
	Type   string `json:"type"`
	NodeID string `json:"nodeid"`
	Path   string `json:"path"`
}

// SubvolReq represents Sub volume Request
type SubvolReq struct {
	Type               string      `json:"type"`
	Bricks             []BrickReq  `json:"bricks"`
	Subvols            []SubvolReq `json:"subvols"`
	ReplicaCount       int         `json:"replica"`
	ArbiterCount       int         `json:"arbiter"`
	DisperseCount      int         `json:"disperse-count,omitempty"`
	DisperseData       int         `json:"disperse-data,omitempty"`
	DisperseRedundancy int         `json:"disperse-redundancy,omitempty"`
}

// VolCreateReq represents a Volume Create Request
type VolCreateReq struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport,omitempty"`
	Subvols   []SubvolReq       `json:"subvols"`
	Options   map[string]string `json:"options,omitempty"`
	Force     bool              `json:"force,omitempty"`
}

// VolOptionReq represents an incoming request to set volume options
type VolOptionReq struct {
	Options map[string]string `json:"options"`
}

// VolExpandReq represents a request to expand the volume by adding more bricks
type VolExpandReq struct {
	ReplicaCount int        `json:"replica,omitempty"`
	Bricks       []BrickReq `json:"bricks"`
	Force        bool       `json:"force,omitempty"`
}

// VolumeOption represents an option that is part of a profile
type VolumeOption struct {
	Name     string `json:"name"`
	OnValue  string `json:"onvalue"`
	OffValue string `json:"offvalue"`
}

// OptionGroup represents a group of options
type OptionGroup struct {
	Name    string         `json:"name"`
	Options []VolumeOption `json:"options"`
}

// OptionGroupReq represents a request to create a new option group
type OptionGroupReq OptionGroup
