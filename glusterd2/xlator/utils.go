package xlator

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/xlator/options"
)

// NotFoundError is returned when an xlator cannot be found in the xlMap
type NotFoundError string

func (e NotFoundError) Error() string {
	return fmt.Sprintf("xlator not found: %s", string(e))
}

// Find returns a Xlator with the give ID if found.
// Returns a XlatorNotFoundError otherwise.
func Find(id string) (*Xlator, error) {
	xl, ok := xlMap[id]
	if !ok {
		return nil, NotFoundError(id)
	}
	return xl, nil
}

// OptionNotFoundError is returned when option with given key cannot be found
// in optMap
type OptionNotFoundError string

func (e OptionNotFoundError) Error() string {
	return fmt.Sprintf("option not found: %s", string(e))
}

// FindOption returns an option.Option for the given key if found.
// key should be in the [<graph>].<xlator>.<name> form.
// Returns an error otherwise.
func FindOption(k string) (*options.Option, error) {
	// Interested only in <xlator>.<name> part of the key as optMap is indexed
	// using them.
	_, xl, name, err := options.SplitKey(k)
	if err != nil {
		return nil, err
	}

	opt, ok := optMap[xl+"."+name]
	if !ok {
		return nil, OptionNotFoundError(k)
	}
	return opt, nil
}

// Contains returns true if string is present in the list else false
func Contains(s string, list []string) bool {
	for _, val := range list {
		if s == val {
			return true
		}
	}
	return false
}
