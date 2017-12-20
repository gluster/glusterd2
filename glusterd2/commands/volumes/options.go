package volumecommands

type option struct {
	OptionName string
	OnValue    string
	OffValue   string
}

// GroupOptions maps from a profile name to a set of options
var GroupOptions = map[string][]option{
	"profile.test": {{"afr.eager-lock", "on", "off"},
		{"gfproxy.afr.eager-lock", "on", "off"}},
}
