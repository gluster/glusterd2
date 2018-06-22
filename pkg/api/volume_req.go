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
/*supported Flags
"reuse-bricks" : for reusing of bricks
"allow-root-dir" : allow root directory to create brick
"allow-mount-as-brick" : reuse if its already mountpoint
"create-brick-dir" : if brick dir is not present, create it
*/
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
	Flags        map[string]bool   `json:"flags,omitempty"`
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
	Options []string `json:"options"`
	Force   bool     `json:"force,omitempty"`
}

// VolExpandReq represents a request to expand the volume by adding more bricks
/*supported Flags
"reuse-bricks" : for reusing of bricks
"allow-root-dir" : allow root directory to create brick
"allow-mount-as-brick" : reuse if its already mountpoint
"create-brick-dir" : if brick dir is not present, create it
*/
type VolExpandReq struct {
	ReplicaCount int             `json:"replica,omitempty"`
	Bricks       []BrickReq      `json:"bricks"`
	Force        bool            `json:"force,omitempty"`
	Flags        map[string]bool `json:"flags,omitempty"`
}

// VolShrinkReq represents a request to remove bricks from a volume
type VolShrinkReq struct {
	Bricks []BrickReq `json:"bricks"`
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

// VolEditReq represents a volume metadata edit request
type VolEditReq struct {
	Metadata       map[string]string `json:"metadata"`
	DeleteMetadata bool              `json:"delete-metadata"`
}

// VolumeStartReq represents a request to start volume
type VolumeStartReq struct {
	ForceStartBricks bool `json:"force-start-bricks,omitempty"`
}

// MetadataSize returns the size of the volume metadata in VolCreateReq
func (v *VolCreateReq) MetadataSize() int {
	return mapSize(v.Metadata)
}

// MetadataSize returns the size of the volume metadata in VolEditReq
func (v *VolEditReq) MetadataSize() int {
	return mapSize(v.Metadata)
}
