package options

type validationFunc func(val interface{}) bool

type optionDetails struct {
	defaultValue    interface{}
	validationFuncs []validationFunc
	help            string
}

var clusterDefaultOptions = map[string]optionDetails{
	"cluster.lookup-unhashed": optionDetails{
		"on",
		[]validationFunc{stringValidation, onOffValidation},
		"This option if set to ON, does a lookup through all the sub-volumes, in case a lookup didn't return any result from the hash subvolume. If set to OFF, it does not do a lookup on the remaining subvolumes."},
}

var volumeDefaultOptions = map[string]optionDetails{
	"features.read-only": optionDetails{
		"off",
		[]validationFunc{stringValidation, onOffValidation},
		"When \"on\", makes a volume read-only. It is turned \"off\" by default."},
}
