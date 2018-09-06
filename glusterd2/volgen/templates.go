package volgen

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"
)

func (tmpl *Template) addEnabledXlator(volinfo *volume.Volinfo, xlist *[]Xlator, xltr Xlator) error {
	var err error
	xlatorEnabled := xltr.isEnabled(volinfo, tmpl.Name)
	if xlatorEnabled || xltr.EnableByOption {
		xl := xltr
		xl.Options = make(map[string]string)
		xl.Options, err = xltr.getOptions(tmpl.Name, volinfo)
		if err != nil {
			return err
		}
		if xltr.EnableByOption {
			if xlatorEnabled {
				xl.Options[xltr.suffix()] = "on"
			} else {
				xl.Options[xltr.suffix()] = "off"
			}
		}

		*xlist = append(*xlist, xl)
	}
	return nil
}

// GetTemplate gets template for the given namespace
func GetTemplate(namespace string, name string) (*Template, error) {
	tmpls, exists := namespaces[namespace]
	if !exists {
		return nil, errors.New("invalid template namespace")
	}

	tmpl, exists := tmpls[name]
	if !exists {
		return nil, errors.New("invalid template name")
	}

	return &tmpl, nil
}

// GetTemplateFromVolinfo gets template from the namespace set in volinfo
// If template namespace is not set in volinfo, gets the template
// from default namespace
func GetTemplateFromVolinfo(volinfo *volume.Volinfo, name string) (*Template, error) {
	tmplNamespace, exists := volinfo.Metadata[TemplateMetadataKey]
	if !exists {
		tmplNamespace = DefaultTemplateNamespace
	}
	return GetTemplate(tmplNamespace, name)
}

// EnabledXlators returns list of xlators which are enabled in Volinfo or in template itself
func (tmpl *Template) EnabledXlators(volinfo *volume.Volinfo) ([]Xlator, error) {
	var (
		xlist []Xlator
		err   error
	)

	for _, xl := range tmpl.Xlators {
		err = tmpl.addEnabledXlator(volinfo, &xlist, xl)
		if err != nil {
			return []Xlator{}, err
		}
	}
	return xlist, nil
}

// EnabledVolumeGraphXlators returns list of xlators enabled in volume level graph
func (tmpl *Template) EnabledVolumeGraphXlators(volinfo *volume.Volinfo) ([]Xlator, error) {
	var (
		xlist []Xlator
		err   error
	)

	for _, xl := range tmpl.VolumeGraphXlators {
		err = tmpl.addEnabledXlator(volinfo, &xlist, xl)
		if err != nil {
			return []Xlator{}, err
		}
	}
	return xlist, nil
}

// EnabledSubvolGraphXlators returns list of xlators enabled in subvol level graph
func (tmpl *Template) EnabledSubvolGraphXlators(volinfo *volume.Volinfo, subvolinfo *volume.Subvol) ([]Xlator, error) {
	var (
		xlist []Xlator
		err   error
	)

	for _, xl := range tmpl.SubvolGraphXlators {
		if xl.TypeTmpl != "" {
			xl.Type = xl.TypeTmpl
			if isVarStr(xl.TypeTmpl) {
				xl.Type, err = varStrReplace(xl.TypeTmpl, utils.MergeStringMaps(
					volinfo.StringMap(),
					subvolinfo.StringMap(),
				))
				if err != nil {
					return []Xlator{}, err
				}
			}
		}
		err = tmpl.addEnabledXlator(volinfo, &xlist, xl)
		if err != nil {
			return []Xlator{}, err
		}
	}
	return xlist, nil
}

// EnabledBrickGraphXlators returns list of xlators enabled in brick level graph
func (tmpl *Template) EnabledBrickGraphXlators(volinfo *volume.Volinfo, subvolinfo *volume.Subvol, brickinfo *brick.Brickinfo) ([]Xlator, error) {
	var (
		xlist []Xlator
		err   error
	)

	for _, xl := range tmpl.BrickGraphXlators {
		if xl.TypeTmpl != "" {
			xl.Type = xl.TypeTmpl
			if isVarStr(xl.TypeTmpl) {
				xl.Type, err = varStrReplace(xl.TypeTmpl, utils.MergeStringMaps(
					volinfo.StringMap(),
					subvolinfo.StringMap(),
					brickinfo.StringMap(),
				))
				if err != nil {
					return []Xlator{}, err
				}
			}
		}
		err = tmpl.addEnabledXlator(volinfo, &xlist, xl)
		if err != nil {
			return []Xlator{}, err
		}
	}
	return xlist, nil
}

// IsXlatorSupported returns true if a xlator is supported for a given template
func (tmpl *Template) IsXlatorSupported(xlatorname string) bool {
	for _, xl := range tmpl.Xlators {
		if xlatorname == xl.fullName(tmpl.Name) || xlatorname == xl.name() || xlatorname == xl.suffix() {
			return true
		}
	}
	return false
}
