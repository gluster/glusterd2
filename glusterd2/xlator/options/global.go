package options

import (
	"github.com/gluster/glusterd2/glusterd2/xlator"
)

// Options is a map of all available xlator options, indexed by
// <xlator.ID>.<option.Key>, for all available option keys.
// Useful for looking up Option during option validation
var Options map[string]*Option

// Load loads all available options into the options.Options map,
// indexed as <xlator.ID>.<option.Key> for all available option keys
func Load() {
	Options = make(map[string]*Option)
	for _, xl := range xlator.Xlators {
		for _, opt := range xl.Options {
			for _, k := range opt.Key {
				k := xl.ID + "." + k
				Options[k] = &Option{opt}
			}
		}
	}
}
