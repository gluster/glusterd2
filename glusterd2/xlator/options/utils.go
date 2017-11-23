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
