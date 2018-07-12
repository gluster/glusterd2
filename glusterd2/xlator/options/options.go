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
	OptionFlagSettable OptionFlag = 1 << iota
	OptionFlagClientOpt
	OptionFlagGlobal
	OptionFlagForce
	OptionFlagNeverReset
	OptionFlagDoc
	// Setting FlagNone instead of the beginning as iota starts incrementing from
	// the first line in a const block, not the first line it is used.
	// Ref: https://github.com/golang/go/wiki/Iota
	OptionFlagNone = 0
)

// OptionLevel is the level at which option is visible to users
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
		return "Undefined"
	}
}

// ErrInvalidArg validates if argument is Invalid
var ErrInvalidArg = errors.New("invalid Value")

// ErrEmptyArg validates for empty arguments
var ErrEmptyArg = errors.New("no value passed")

//ErrInvalidRange validates if option is out of range
var ErrInvalidRange = errors.New("option is out of valid range")

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
	Flags        OptionFlag
	Tags         []string
	SetKey       string
	Level        OptionLevel
}

// IsSettable returns true if the option can be set by a user, returns false
// otherwise.
func (o *Option) IsSettable() bool {
	return (o.Flags & OptionFlagSettable) == OptionFlagSettable
}

// IsNeverReset returns true if the option should never be set by a user,
// returns false otherwise.
func (o *Option) IsNeverReset() bool {
	return (o.Flags & OptionFlagNeverReset) == OptionFlagNeverReset
}

// IsForceRequired returns true if the option requires force variable for the
// user to set returns false otherwise.
func (o *Option) IsForceRequired() bool {
	return (o.Flags & OptionFlagForce) == OptionFlagForce
}

// IsAdvanced returns true if the option is an advanced option
func (o *Option) IsAdvanced() bool {
	return o.Level == OptionStatusAdvanced
}

// IsExperimental returns true if the option is experimental
func (o *Option) IsExperimental() bool {
	return o.Level == OptionStatusExperimental
}

// IsDeprecated returns true if the option is deprcated
func (o *Option) IsDeprecated() bool {
	return o.Level == OptionStatusDeprecated
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
		err = ValidateInternetAddressList(o, val)
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
	if len(val) == 0 {
		return ErrInvalidArg
	}
	if validate.IsHost(val) {
		return nil
	} else if validate.IsIP(val) {
		return nil
	} else if validate.IsCIDR(val) {
		return nil
	} else if strings.ContainsAny(val, "* & # & ? & ^") {
		return nil
	}

	return ErrInvalidArg
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
// Example: val := "k1:1024KB, k2:512MB, k3:512GB"
// It is verified next if 1024KB is a valid size.
func ValidatePriorityOrSize(o *Option, val string) error {
	pairs := strings.Split(val, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if strings.HasSuffix(kv[1], "B") || strings.HasSuffix(kv[1], "b") {
			err := ValidateSizeList(o, kv[1])
			return err
		}
		_, err := strconv.ParseInt(kv[1], 10, 64)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateSizeList validates if the option is a valid size list
func ValidateSizeList(o *Option, val string) error {
	var sizeinbytes int64
	slist := strings.Split(val, ",")
	for _, el := range slist {
		el = strings.TrimSpace(el)
		l := len(el)
		if strings.HasSuffix(el, "B") || strings.HasSuffix(el, "b") {
			size := el[l-2 : l]
			v, err := strconv.ParseInt(el[:l-2], 10, 64)
			if err != nil {
				return err
			}
			switch size {
			case "KB", "kb":
				sizeinbytes = v * 1024
			case "MB", "mb":
				sizeinbytes = v * 1024 * 1024
			case "GB", "gb":
				sizeinbytes = v * 1024 * 1024 * 1024
			}
			if sizeinbytes%512 != 0 {
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
	var time int    // to convert given value from other formates to seconds
	var tStr string // temp value which has the int without "min" ...
	multiplier := 1
	if strings.HasSuffix(val, "sec") {
		tStr = strings.TrimSuffix(val, "sec")
	} else if strings.HasSuffix(val, "s") {
		tStr = strings.TrimSuffix(val, "s")
	} else if strings.HasSuffix(val, "min") {
		tStr = strings.TrimSuffix(val, "min")
		multiplier = 60
	} else if strings.HasSuffix(val, "m") {
		tStr = strings.TrimSuffix(val, "m")
		multiplier = 60
	} else if strings.HasSuffix(val, "hr") {
		tStr = strings.TrimSuffix(val, "hr")
		multiplier = 60 * 60
	} else if strings.HasSuffix(val, "h") {
		tStr = strings.TrimSuffix(val, "h")
		multiplier = 60 * 60
	} else if strings.HasSuffix(val, "days") {
		tStr = strings.TrimSuffix(val, "days")
		multiplier = 60 * 60 * 24
	} else if strings.HasSuffix(val, "d") {
		tStr = strings.TrimSuffix(val, "d")
		multiplier = 60 * 60 * 24
	} else if strings.HasSuffix(val, "w") {
		tStr = strings.TrimSuffix(val, "w")
		multiplier = 60 * 60 * 24 * 7
	} else if strings.HasSuffix(val, "wk") {
		tStr = strings.TrimSuffix(val, "wk")
		multiplier = 60 * 60 * 24 * 7
	} else {
		tStr = val
	}
	time, err := strconv.Atoi(tStr)
	if err != nil {
		return err
	}
	time = time * multiplier
	return ValidateRange(o, strconv.Itoa(time))
}

// ValidateXlator validates if the option is a valid xlator
func ValidateXlator(o *Option, val string) error {
	return ErrInvalidArg
}

// ValidateOption validates if the option is valid
func ValidateOption(o *Option, val string) error {
	return ErrInvalidArg
}

// StringToBoolean converts probable boolean strings to True or False
func StringToBoolean(val string) (bool, error) {
	if val == "" {
		return false, ErrEmptyArg
	}

	switch strings.ToLower(val) {
	case "on", "yes", "true", "enable", "1":
		return true, nil
	case "off", "no", "false", "disable", "0":
		return false, nil
	default:
		return false, ErrInvalidArg
	}
}
