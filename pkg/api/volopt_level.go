package api

// OptionLevel is the level at which option is visible to users
//go:generate jsonenums -type=OptionLevel
type OptionLevel uint

// These are the available option levels
const (
	OptionStatusAdvanced OptionLevel = iota
	OptionStatusBasic
	OptionStatusExperimental
	OptionStatusDeprecated
)

func (l OptionLevel) String() string {
	switch l {
	case OptionStatusBasic:
		return "Basic"
	case OptionStatusAdvanced:
		return "Advanced"
	case OptionStatusExperimental:
		return "Experimental"
	case OptionStatusDeprecated:
		return "Deprecated"
	default:
		return "invalid option level"
	}
}
