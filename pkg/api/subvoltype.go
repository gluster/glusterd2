package api

// SubvolType is the Type of the volume
//go:generate jsonenums -type=SubvolType
type SubvolType uint16

const (
	// SubvolDistribute is a distribute sub volume
	SubvolDistribute SubvolType = iota
	// SubvolReplicate is a replicate sub volume
	SubvolReplicate
	// SubvolDisperse is a disperse sub volume
	SubvolDisperse
)
