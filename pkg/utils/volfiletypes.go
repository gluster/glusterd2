package utils

const (
	// BrickVolfile is name of brick volfile template
	BrickVolfile = "brick"
	// ClientVolfile is a name of client volfile template
	ClientVolfile = "client"
	// RebalanceVolfile is a name of rebalance volfile template
	RebalanceVolfile = "rebalance"
	// SelfHealVolfile is a name of selfheal volfile template
	SelfHealVolfile = "glustershd"
	// BitdVolfile is a name of bitd volfile template
	BitdVolfile = "bitd"
	// ScrubdVolfile is a name of scrubd volfile template
	ScrubdVolfile = "scrubd"
	// GfProxyVolfile is a name of gfproxy volfile template
	GfProxyVolfile = "gfproxy"
	// NFSVolfile is a name of nfs volfile template
	NFSVolfile = "nfs"
)

// ValidVolfiles represents list of valid volfile names
var ValidVolfiles = [...]string{
	BrickVolfile,
	ClientVolfile,
	RebalanceVolfile,
	SelfHealVolfile,
	BitdVolfile,
	ScrubdVolfile,
	GfProxyVolfile,
	NFSVolfile,
}
