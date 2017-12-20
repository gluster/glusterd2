package options

import (
	"errors"
	"net"
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

// ErrInvalidArg validates if argument is Invalid
var ErrInvalidArg = errors.New("Invalid Value")

// ErrEmptyArg validates for empty arguments
var ErrEmptyArg = errors.New("No value passed")

//ErrInvalidRange validates if option is out of range
var ErrInvalidRange = errors.New("Option is out of valid range")

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
	switch o.Type {
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

// ValidateRange validates if option in correctrange.
func ValidateRange(o *Option, val string) error {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	if o.ValidateType == OptionValidateBoth && o.Min == 0 && o.Max == 0 {
		return nil
	} else if o.ValidateType == OptionValidateMin && v < o.Min {
		return ErrInvalidRange
	} else if o.ValidateType == OptionValidateMax && v > o.Max {
		return ErrInvalidRange
	} else if v < o.Min || v > o.Max {
		return ErrInvalidRange
	}
	return nil
}

// ValidateBool validates if the option is of type boolean
func ValidateBool(o *Option, val string) error {
	if val == "" {
		return ErrEmptyArg
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
	var err error
	if validate.IsFloat(val) != true {
		err = ErrInvalidArg
	} else {
		err = ValidateRange(o, val)
	}
	return err
}

// ValidateInt validates if the option is of type Int
func ValidateInt(o *Option, val string) error {
	var err error
	if validate.IsInt(val) != true {
		err = ErrInvalidArg
	} else {
		err = ValidateRange(o, val)
	}
	return err
}

// ValidateInternetAddress validates the Internet Address
func ValidateInternetAddress(o *Option, val string) error {
	_, _, err := net.ParseCIDR(val)
	if err != nil {
		return ErrInvalidArg
	} else if validate.IsIP(val) != true {
		return ErrInvalidArg
	} else if validate.IsHost(val) != true {
		return ErrInvalidArg
	} else if strings.ContainsAny(val, "* & # & ? & ^") == true {
		return ErrInvalidArg
	}
	return nil
}

// ValidateInternetAddressList validates the Internet Address List
func ValidateInternetAddressList(o *Option, val string) error {
	iplist := strings.Split(val, ",")
	for _, ip := range iplist {
		err := ValidateInternetAddress(o, ip)
		if err != nil {
			return ErrInvalidArg
		}
	}
	return nil
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
		err = ErrInvalidRange
	}
	return err
}

// ValidatePercentOrSize validates either a correct percent format or size
func ValidatePercentOrSize(o *Option, val string) error {
	var err error
	if strings.ContainsRune(val, '%') {
		err = ValidatePercent(o, val)
	} else {
		err = ValidateSizet(o, val)
	}
	return err
}

// ValidatePriorityOrSize validates either priority or size
func ValidatePriorityOrSize(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateSizeList validates if the option is a valid size list
func ValidateSizeList(o *Option, val string) error {
	slist := strings.Split(val, ",")
	for _, el := range slist {
		l := len(el)
		if strings.ContainsRune(el, 'B') || strings.ContainsRune(el, 'b') {
			t := strings.TrimSpace(el[:l-2])
			v, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				return err
			} else if v%512 != 0 {
				return ErrInvalidArg
			}
		}
	}
	return nil
}

// ValidateSizet validates if the option is a valid size
func ValidateSizet(o *Option, val string) error {
	err := ValidateRange(o, val)
	return err
}

// ValidateStr validates if the option is of type Str
func ValidateStr(o *Option, val string) error {
	t := strings.TrimSpace(val)
	l := len(t)
	if l == 0 {
		return ErrEmptyArg
	}
	if len(o.Value) == 0 {
		return nil
	}
	for _, op := range o.Value {
		if t == op {
			return nil
		}
		return ErrInvalidArg
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
