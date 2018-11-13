package volgen

import (
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/utils"
)

func (xl *Xlator) isEnabled(volinfo *volume.Volinfo, tmplName string) bool {
	// If Volinfo.Options can contains xlator, check for xlator
	// existence in the following order: FullName, Name and Suffix
	// For example changelog can be enabled using
	// `brick.features.changelog` or `features.changelog`
	// or `changelog`
	// volinfo can be nil in case of cluster level graphs
	if volinfo != nil {
		value, exists := volinfo.Options[xl.fullName(tmplName)]

		if !exists {
			value, exists = volinfo.Options[xl.name()]
		}

		if !exists {
			value, exists = volinfo.Options[xl.suffix()]
		}
		if exists {
			return boolify(value)
		}
	}

	// If xlator is not enabled in volinfo.Options then take
	// the default enabled value from template
	return !xl.Disabled
}

func (xl *Xlator) fullName(tmplName string) string {
	return tmplName + "." + xl.Type
}

func (xl *Xlator) name() string {
	return xl.Type
}

func (xl *Xlator) suffix() string {
	return path.Base(xl.Type)
}

func (xl *Xlator) getOptions(tmplName string, volinfo *volume.Volinfo) (map[string]string, error) {
	var optKey, optVal string

	opts := make(map[string]string)
	// Load default options
	xlid := xl.suffix()
	xlopts, err := xlator.Find(xlid)
	if err != nil {
		return nil, err
	}

	for _, o := range xlopts.Options {
		// If the option has an explicit SetKey, use it as the key
		optKey = o.Key[0]
		if o.SetKey != "" {
			optKey = o.SetKey
		}

		if utils.StringInSlice(optKey, xl.IgnoreOptions) {
			continue
		}

		optVal = o.DefaultValue

		// If option set in template
		v, ok := xl.Options[optKey]
		if ok {
			opts[optKey] = v
		}

		// Volinfo can be nil in case of cluster level
		if volinfo == nil {
			opts[optKey] = optVal
			continue
		}

		// Special case: Option may be set for any key, iterate
		// for each key and check if option is set in volinfo
		// for any of the keys
		for _, k := range o.Key {
			// If option set as <template-name>.<xlid>.<option>
			// For example: client.io-stats.log-level
			v, ok := volinfo.Options[tmplName+"."+xlid+"."+k]
			if ok {
				optVal = v
				break
			}
			// If option set as <template-name>.<xltype>/<xlid>.<option>
			// For example: client.debug/io-stats.log-level
			v, ok = volinfo.Options[tmplName+"."+xl.Type+"."+k]
			if ok {
				optVal = v
				break
			}
			// If option set as <xlid>.<option>
			// For example: io-stats.log-level
			v, ok = volinfo.Options[xlid+"."+k]
			if ok {
				optVal = v
				break
			}
			// If option set as <xltype>/<xlid>.<option>
			// For example: debug/io-stats.log-level
			v, ok = volinfo.Options[xl.Type+"."+k]
			if ok {
				optVal = v
				break
			}
		}
		opts[optKey] = optVal
	}

	// Template options are already substituted in opts. If any
	// option is set in template which is not part of options table
	// then add that.(May be virtual option related to enable/disable)
	for k, v := range xl.Options {
		_, ok := opts[k]
		if !ok {
			opts[k] = v
		}
	}

	// Do not try to substitute from volinfo in case of cluster level
	if volinfo == nil {
		return opts, nil
	}

	// First iteration look for option names set without template name
	for k, v := range volinfo.Options {
		if strings.HasPrefix(k, xlid+".") || strings.HasPrefix(k, xl.Type+".") {
			parts := strings.Split(k, ".")
			k = parts[len(parts)-1]
			if !utils.StringInSlice(k, xl.IgnoreOptions) {
				opts[k] = v
			}
		}
	}

	// Second iteration look for option names set with template name
	// so that option set with template name takes precedence
	for k, v := range volinfo.Options {
		if strings.HasPrefix(k, tmplName+"."+xlid+".") || strings.HasPrefix(k, tmplName+"."+xl.Type+".") {
			parts := strings.Split(k, ".")
			k = parts[len(parts)-1]
			if !utils.StringInSlice(k, xl.IgnoreOptions) {
				opts[k] = v
			}
		}
	}

	return opts, nil
}
