package xlator

// OptionType is a type which represents the type of xlator option.
type OptionType int

// These are all the possile values for the type OptionType.
const (
	OptionTypeAny OptionType = iota
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

// OptionValidateType is a type which represents how the value of xlator
// option should be validated.
type OptionValidateType int

// These are all the possile values for the type OptionValidateType
const (
	OptionValidateBoth OptionValidateType = iota
	OptionValidateMin
	OptionValidateMax
)

type OptionFlag uint

const (
	OptionFlagNone     = 0
	OptionFlagSettable = 1 << iota
	OptionFlagClientOpt
	OptionFlagGlobal
	OptionFlagForce
	OptionFlagNeverReset
	OptionFlagDoc
)

// Option is a struct which represents one single xlator option exported by
// the translator.
type Option struct {
	Key          []string
	Type         OptionType
	Value        []string
	DefaultValue string
	Description  string
	Min          float64
	Max          float64
	Validate     OptionValidateType
	OpVersion    []uint32
	Deprecated   []uint32
	Flags        uint32
	Tags         []string
}
