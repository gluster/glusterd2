package options

import (
	"github.com/gluster/glusterd2/pkg/types"
)

// These are all the possile values for the type OptionType.
const (
	OptionTypeAny types.OptionType = iota
	OptionTypeStr
	OptionTypeInt
	OptionTypeSizet
	OptionTypePercent
	OptionTypePercentOrSizet
	OptionTypeBool
	OptionTypeXlator
	OptionTypePath
	OptionTypeTime
	OptionTypeDouble
	OptionTypeInternetAddress
	OptionTypeInternetAddressList
	OptionTypePriorityList
	OptionTypeSizeList
	OptionTypeClientAuthAddr
)

// These are all the possile values for the type OptionValidateType
const (
	OptionValidateBoth types.OptionValidateType = iota
	OptionValidateMin
	OptionValidateMax
)

// These are the available types.OptionFlags
const (
	OptionFlagNone     types.OptionFlag = 0
	OptionFlagSettable                  = 1 << iota
	OptionFlagClientOpt
	OptionFlagGlobal
	OptionFlagForce
	OptionFlagNeverReset
	OptionFlagDoc
)

// Option is a struct which represents one single xlator option exported by
// the translator.
// Embedding the actual type declared in package types, to allow custom methods.
type Option struct {
	*types.Option
}

// Validate checks if the given value string can be set as the value for the
// Option.
// Returns are error if it is not possible, nil otherwise.
func (o *Option) Validate(val string) error {
	// TODO: Do actual validation
	return nil
}
