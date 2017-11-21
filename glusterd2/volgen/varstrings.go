package volgen

import (
	"fmt"
	"regexp"
	"strings"
)

var varStrRE = regexp.MustCompile(`\{\{\s*(\S+)\s*\}\}`)

// UnknownVarStrErr is returned when a varstring is not found in the given map
type UnknownVarStrErr string

func (e UnknownVarStrErr) Error() string {
	return fmt.Sprintf("unknown variable string: %s", string(e))
}

func isVarStr(s string) bool {
	return varStrRE.MatchString(s)
}

func varStr(s string) string {
	return strings.Trim(varStrRE.FindString(s), "{} ")
}

func varStrReplace(s string, vals map[string]string) (string, error) {
	k := varStr(s)
	v, ok := vals[k]
	if !ok {
		return "", UnknownVarStrErr(k)
	}
	return varStrRE.ReplaceAllString(s, v), nil
}
