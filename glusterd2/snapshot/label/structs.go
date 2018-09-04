//Package label that contains struct for label
package label

//Options is used as key to set Label info values
type Options string

const (
	//SnapMaxHardLimitKey is used to set the value SnapMaxHardLimit
	SnapMaxHardLimitKey Options = "snap-max-hard-limit"
	//SnapMaxSoftLimitKey is used to set the value SnapMaxSoftLimit
	SnapMaxSoftLimitKey Options = "snap-max-soft-limit"
	//ActivateOnCreateKey is used to set the value ActivateOnCreate
	ActivateOnCreateKey Options = "activate-on-create"
	//AutoDeleteKey is used to set the value AutoDelete
	AutoDeleteKey Options = "auto-delete"
)

//Info is used to represent a label
type Info struct {
	Name             string
	SnapMaxHardLimit uint64
	SnapMaxSoftLimit uint64
	ActivateOnCreate bool
	AutoDelete       bool
	Description      string
	SnapList         []string
}

//DefaultLabel contains default values for a label
var DefaultLabel = Info{
	ActivateOnCreate: false,
	AutoDelete:       false,
	Description:      "This is a default label",
	Name:             "defaultLabel",
	SnapMaxHardLimit: 256,
	SnapMaxSoftLimit: 230,
}
