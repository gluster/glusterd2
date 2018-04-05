package api

// BrickReq represents Brick Request
type BrickReq struct {
	Type   string `json:"type"`
	PeerID string `json:"peerid"`
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
	Name         string            `json:"name"`
	Transport    string            `json:"transport,omitempty"`
	Subvols      []SubvolReq       `json:"subvols"`
	Options      map[string]string `json:"options,omitempty"`
	Force        bool              `json:"force,omitempty"`
	Advanced     bool              `json:"advanced,omitempty"`
	Experimental bool              `json:"experimental,omitempty"`
	Deprecated   bool              `json:"deprecated,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// VolOptionReq represents an incoming request to set volume options
type VolOptionReq struct {
	Options      map[string]string `json:"options"`
	Advanced     bool              `json:"advanced,omitempty"`
	Experimental bool              `json:"experimental,omitempty"`
	Deprecated   bool              `json:"deprecated,omitempty"`
}

// VolOptionResetReq represents a request to reset volume options
type VolOptionResetReq struct {
	Options map[string]string `json:"options"`
	Force   bool              `json:"force,omitempty"`
}

// VolExpandReq represents a request to expand the volume by adding more bricks
type VolExpandReq struct {
	ReplicaCount int        `json:"replica,omitempty"`
	Bricks       []BrickReq `json:"bricks"`
	Force        bool       `json:"force,omitempty"`
}

// VolumeOption represents an option that is part of a profile
type VolumeOption struct {
	Name    string `json:"name"`
	OnValue string `json:"onvalue"`
}

// OptionGroup represents a group of options
type OptionGroup struct {
	Name        string         `json:"name"`
	Options     []VolumeOption `json:"options"`
	Description string         `json:"description"`
}

// OptionGroupReq represents a request to create a new option group
type OptionGroupReq struct {
	OptionGroup
	Advanced     bool `json:"advanced,omitempty"`
	Experimental bool `json:"experimental,omitempty"`
	Deprecated   bool `json:"deprecated,omitempty"`
}

// ClientStatedump uniquely identifies a client (only gfapi) connected to
// glusterd2
type ClientStatedump struct {
	Host string `json:"host" valid:"host,required"`
	Pid  int    `json:"pid" valid:"required"`
}

// VolStatedumpReq represents a request to take statedump of various processes
// of a volume
type VolStatedumpReq struct {
	Bricks bool            `json:"bricks,omitempty"`
	Quota  bool            `json:"quotad,omitempty"`
	Client ClientStatedump `json:"client,omitempty"`
}
