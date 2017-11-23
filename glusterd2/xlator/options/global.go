package options

// Options is a map of all available xlator options, indexed by
// <xlator.ID>.<option.Key>, for all available option keys.
// Useful for looking up Option during option validation
var Options map[string]*Option

func init() {
	Options = make(map[string]*Option)
}
