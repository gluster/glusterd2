package api

// VolState is the current state of the volume.
//go:generate jsonenums -type=VolState
type VolState uint16

const (
	// VolCreated represents a volume that has been just created and never started.
	VolCreated VolState = iota
	// VolStarted represents a volume in started state.
	VolStarted
	// VolStopped represents a volume in stopped state (excluding newly created but never started volumes).
	VolStopped
)

func (s VolState) String() string {
	switch s {
	case VolCreated:
		return "Created"
	case VolStarted:
		return "Started"
	case VolStopped:
		return "Stopped"
	default:
		return "invalid VolState"
	}
}
