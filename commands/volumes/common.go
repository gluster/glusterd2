package volumecommands

import (
	"strings"

	"github.com/gluster/glusterd2/xlator"
)

func areOptionNamesValid(optsFromReq map[string]string) bool {

	var xlOptFound bool
	for o := range optsFromReq {

		// assuming option to be of the form <domain>.<xlator-option>
		// and <domain> will be the xlator type.
		// Example: cluster/afr.eager-lock
		// we know for certain that this isn't true

		tmp := strings.Split(strings.TrimSpace(o), ".")
		if len(tmp) != 2 {
			return false
		}
		xlatorType := tmp[0]
		xlatorOption := tmp[1]

		options, ok := xlator.AllOptions[xlatorType]
		if !ok {
			return false
		}

		xlOptFound = false
		for _, option := range options {
			for _, key := range option.Key {
				if xlatorOption == key {
					xlOptFound = true
				}
			}
		}
		if !xlOptFound {
			return false
		}
	}

	return true
}
