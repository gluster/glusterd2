package volgen

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/utils"
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

// getValue returns value if found for provided graph.xlator.keys in the options map
// XXX: Not possibly the best place for this
func getValue(graph, xl string, keys []string, opts map[string]string) (string, string, bool) {
	for _, k := range keys {
		v, ok := opts[graph+"."+xl+"."+k]
		if ok {
			return k, v, true
		}
		v, ok = opts[xl+"."+k]
		if ok {
			return k, v, true
		}
	}

	return "", "", false
}

func (e *Entry) getOptions(graph string, extra *map[string]extrainfo) (map[string]string, error) {
	var (
		xl  *xlator.Xlator
		err error
	)

	xlid := path.Base(e.Type)
	xl, err = xlator.Find(xlid)
	if err != nil {
		return nil, err
	}

	opts := make(map[string]string)
	if e.VolumeID != nil {
		opts = (*extra)[e.VolumeID.String()].Options
	}

	data := make(map[string]string)
	if e.VolumeID != nil {
		key := e.VolumeID.String()

		if e.BrickID != nil {
			key += "." + e.BrickID.String()
		}
		data = (*extra)[e.VolumeID.String()].StringMaps[key]
	}

	data = utils.MergeStringMaps(data, e.ExtraData)

	xlopts := make(map[string]string)

	for _, o := range xl.Options {
		var (
			k, v string
			ok   bool
		)

		// If the option has an explicit SetKey, use it as the key
		if o.SetKey != "" {
			k = o.SetKey
			_, v, ok = getValue(graph, xlid, o.Key, opts)
		} else {
			k, v, ok = getValue(graph, xlid, o.Key, opts)
		}

		// If the option is not found in Volinfo, try to set to defaults if
		// available and required
		if !ok {
			// If there is no default value skip setting this option
			if o.DefaultValue == "" {
				continue
			}
			v = o.DefaultValue

			if k == "" {
				k = o.Key[0]
			}
		}

		// Do varsting replacements if required
		keyChanged := false
		keyVarStr := isVarStr(k)
		if keyVarStr {
			k1, err := varStrReplace(k, data)
			if err != nil {
				return nil, err
			}
			if k != k1 {
				keyChanged = true
			}
			k = k1
		}
		if isVarStr(v) {
			if v, err = varStrReplace(v, data); err != nil {
				return nil, err
			}
		}
		// Set the option
		// Ignore setting if value is empty or if key is
		// varstr and not changed after substitute. This can happen
		// only if the field value is empty
		if v != "" || (keyVarStr && !keyChanged) {
			xlopts[k] = v
		}
	}

	// Set all the extra Options
	for k, v := range e.ExtraOptions {
		xlopts[k] = v
	}

	return xlopts, nil
}
