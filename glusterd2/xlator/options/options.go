package options

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	validate "github.com/asaskevich/govalidator"
)

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

// OptionFlag is the type representing the flags of an Option
type OptionFlag uint

// These are the available OptionFlags
const (
	OptionFlagNone     = 0
	OptionFlagSettable = 1 << iota
	OptionFlagClientOpt
	OptionFlagGlobal
	OptionFlagForce
	OptionFlagNeverReset
	OptionFlagDoc
)

// ErrInvalidArg is an Invalid Argument error
var ErrInvalidArg = errors.New("Invalid Argument")

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

// Validate checks if the given value string can be set as the value for the
// Option.
// Returns are error if it is not possible, nil otherwise.
func (o *Option) Validate(val string) error {
	var err error
	switch t := o.Type; t {
	case OptionTypeBool:
		err = ValidateBool(o, val)
	case OptionTypeClientAuthAddr:
		err = ValidateClientAuthAddr(o, val)
	case OptionTypeDouble:
		err = ValidateDouble(o, val)
	case OptionTypeInt:
		err = ValidateInt(o, val)
	case OptionTypeInternetAddress:
		err = ValidateInternetAddress(o, val)
	case OptionTypeInternetAddressList:
		err = ValidateInternetAddressList(o, val)
	case OptionTypePath:
		err = ValidatePath(o, val)
	case OptionTypePercent:
		err = ValidatePercent(o, val)
	case OptionTypePercentOrSizet:
		err = ValidatePercentOrSize(o, val)
	case OptionTypePriorityList:
		err = ValidatePriorityOrSize(o, val)
	case OptionTypeSizeList:
		err = ValidateSizeList(o, val)
	case OptionTypeSizet:
		err = ValidateSizet(o, val)
	case OptionTypeStr:
		err = ValidateStr(o, val)
	case OptionTypeTime:
		err = ValidateTime(o, val)
	case OptionTypeXlator:
		err = ValidateXlator(o, val)
	default:
		err = ValidateOption(o, val)
	}
	return err
}

// ValidateBool validates if the option is of type boolean
func ValidateBool(o *Option, val string) error {
	if val == "" {
		err := errors.New("No argument passed")
		return err
	}
	switch strings.ToLower(val) {
	case "on", "yes", "true", "enable", "0", "off", "no", "false", "disable", "1":
		return nil
	default:
		return ErrInvalidArg
	}
}

// ValidateClientAuthAddr validates mount auth address
func ValidateClientAuthAddr(o *Option, val string) error {
	return ErrInvalidArg
}

//ValidateDouble validates if the option is of type double
func ValidateDouble(o *Option, val string) error {
	if validate.IsFloat(val) != true {
		return ErrInvalidArg
	}
	return nil
}

// ValidateInt validates if the option is of type Int
func ValidateInt(o *Option, val string) error {
	if validate.IsInt(val) != true {
		return ErrInvalidArg
	}
	return nil
}

// ValidateInternetAddress validates the Internet Address
func ValidateInternetAddress(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateInternetAddressList validates the Internet Address List
func ValidateInternetAddressList(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidatePath validates if the option is a valid path
func ValidatePath(o *Option, val string) error {
	var err error
	t, _ := validate.IsFilePath(val)
	if t != true {
		err = errors.New("invalid path given")
	} else if filepath.IsAbs(val) != true {
		err = errors.New("option is not an absolute path name")
	}
	return err
}

// ValidatePercent validates if the option is in correct percent format
func ValidatePercent(o *Option, val string) error {
	var percent float64
	var err error
	l := len(val)
	if val[0] == '%' {
		err = ErrInvalidArg
	} else if val[l-1] == '%' {
		percent, err = strconv.ParseFloat(val[:(l-1)], 64)
	} else {
		percent, err = strconv.ParseFloat(val, 64)
	}
	if percent < 0.0 || percent > 100.0 {
		err = errors.New("option is out of range [0 - 100]")
	}
	return err
}

// ValidatePercentOrSize validates either a correct percent format or size
func ValidatePercentOrSize(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidatePriorityOrSize validates either priority or size
func ValidatePriorityOrSize(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateSizeList validates if the option is a valid size list
func ValidateSizeList(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateSizet validates if the option is a valid size
func ValidateSizet(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateStr validates if the option is of type Str
func ValidateStr(o *Option, val string) error {
	l := len(val)
	if l == 0 {
		return ErrInvalidArg
	} else if len(strings.TrimSpace(val)) == 0 {
		return ErrInvalidArg
	} else {
		for _, op := range o.Value {
			if val != op {
				return ErrInvalidArg
			}
		}
		return nil
	}
	return nil
}

// ValidateTime validates if the option is valid time format
func ValidateTime(o *Option, val string) error {
	if validate.IsTime(val, "hh:mm:ss") != true {
		return ErrInvalidArg
	}
	return nil
}

// ValidateXlator validates if the option is a valid xlator
func ValidateXlator(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateOption validates if the option is valid
func ValidateOption(o *Option, val string) error {
	return ErrInvalidArg
}
