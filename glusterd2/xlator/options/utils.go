package options

import (
	"fmt"
	"strings"
)

// InvalidKeyError is returned by SplitKey if key is not of the correct form.
type InvalidKeyError string

func (e InvalidKeyError) Error() string {
	return fmt.Sprintf("option key not in <graph>.<xlator>.<name> form: %s", string(e))
}

// OptionNotFoundError is returned when option with given key cannot be found
// in the Options map
type OptionNotFoundError string

func (e OptionNotFoundError) Error() string {
	return fmt.Sprintf("option not found: %s", string(e))
}

// SplitKey returns three strings by breaking key of the form
// [<graph>].<xlator>.<name> into its constituents. <graph> is optional.
// Returns an InvalidKeyError if key is not of correcf form.
func SplitKey(k string) (string, string, string, error) {
	tmp := strings.Split(strings.TrimSpace(k), ".")
	switch len(tmp) {
	case 2:
		return "", tmp[0], tmp[1], nil
	case 3:
		return tmp[0], tmp[1], tmp[2], nil
	default:
		return "", "", "", InvalidKeyError(k)
	}
}

// Find returns an Option for the given key if found.
// key should be in the [<graph>].<xlator>.<name> form.
// Returns an OptionNotFoundError otherwise.
func Find(k string) (*Option, error) {
	// Intersted only in <xlator>.<name> part of the key as Options is indexed
	// using them.
	_, xl, name, err := SplitKey(k)
	if err != nil {
		return nil, err
	}

	opt, ok := Options[xl+"."+name]
	if !ok {
		return nil, OptionNotFoundError(k)
	}
	return opt, nil
}
