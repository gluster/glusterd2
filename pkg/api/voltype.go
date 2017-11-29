package api

// VolType is the type of volume.
//go:generate jsonenums -type=VolType
type VolType uint16

const (
	// Distribute is a plain distribute volume
	Distribute VolType = iota
	// Replicate is plain replicate volume
	Replicate
	// Disperse is a plain erasure coded volume
	Disperse
	// DistReplicate is a distribute-replicate volume
	DistReplicate
	// DistDisperse is a distribute-'erasure coded' volume
	DistDisperse
)

func (t VolType) String() string {
	switch t {
	case Distribute:
		return "Distribute"
	case Replicate:
		return "Replicate"
	case Disperse:
		return "Disperse"
	case DistReplicate:
		return "Distributed-Replicate"
	case DistDisperse:
		return "Distributed-Disperse"
	default:
		return "invalid VolState"
	}
}
