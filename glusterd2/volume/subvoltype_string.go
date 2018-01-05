package volume

func (i SubvolType) String() string {
	switch i {
	case SubvolDistribute:
		return "Distribute"
	case SubvolReplicate:
		return "Replicate"
	case SubvolDisperse:
		return "Disperse"
	default:
		return "Distribute"
	}
}
