package types

// OptionType is a type which represents the type of xlator option.
type OptionType int

// OptionValidateType is a type which represents how the value of xlator
// option should be validated.
type OptionValidateType int

// OptionFlag is the type representing the flags of an Option
type OptionFlag uint

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
	ValidateType OptionValidateType
	OpVersion    []uint32
	Deprecated   []uint32
	Flags        uint32
	Tags         []string
	SetKey       string
}
