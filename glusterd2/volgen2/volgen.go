package volgen2

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
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

// Entry represents one Xlator entry in Volfile
type Entry struct {
	Name         string
	Type         string
	VolumeID     uuid.UUID
	BrickID      uuid.UUID
	ExtraOptions map[string]string
	ExtraData    map[string]string
	SubEntries   []Entry
}

// Volfile represents Gluster Volfile
type Volfile struct {
	Name      string
	FileName  string
	RootEntry Entry
}

// New initializes Entry structure
func New(name string) *Volfile {
	return &Volfile{Name: name, RootEntry: Entry{}}
}

// Add adds sub entry
func (e *Entry) Add(xlatorType string, vol *volume.Volinfo, b *brick.Brickinfo) *Entry {
	name := ""
	var volid uuid.UUID
	var brickid uuid.UUID

	if vol != nil {
		volid = vol.ID
		name = vol.Name + "-" + path.Base(xlatorType)
	}

	if b != nil {
		brickid = b.ID
	}

	e.SubEntries = append(e.SubEntries, Entry{Name: name, VolumeID: volid, BrickID: brickid, Type: xlatorType})

	// Return the last element's pointer, useful if sub elements to be added to the newly added element
	return &e.SubEntries[len(e.SubEntries)-1]
}

// SetName sets entry name
func (e *Entry) SetName(name string) *Entry {
	e.Name = name
	return e
}

// SetExtraOptions sets extra options
func (e *Entry) SetExtraOptions(opts map[string]string) *Entry {
	e.ExtraOptions = opts
	return e
}

// SetExtraData sets extra data
func (e *Entry) SetExtraData(data map[string]string) *Entry {
	e.ExtraData = data
	return e
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

			// If neither key nor value is a varstring, skip setting this option
			if !isVarStr(k) && !isVarStr(v) {
				continue
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

// Generate generates Volfile content
func (v *Volfile) Generate(graph string, extra *map[string]extrainfo) (string, error) {
	if v.Name == "" || v.FileName == "" {
		return "", errors.New("Incomplete details")
	}

	return v.RootEntry.Generate(graph, extra)

}

// Generate generates the Volfile content
func (e *Entry) Generate(graph string, extra *map[string]extrainfo) (string, error) {
	if graph == "" {
		graph = e.Name
	}

	out := ""
	subvolumes := []string{}
	for _, entry := range e.SubEntries {
		out1, err := entry.Generate(graph, extra)
		if err != nil {
			return "", err
		}
		out += out1
		subvolumes = append(subvolumes, entry.Name)
	}

	if e.Type == "" {
		return out, nil
	}

	// volume <name>
	out += "volume " + e.Name + "\n"

	// type <type>
	out += "    type " + e.Type + "\n"

	// option <key> <value>
	// ty := path.Base(e.Type)
	opts, err := e.getOptions(graph, extra)
	if err != nil {
		return "", err
	}
	for k, v := range opts {
		out += "    option " + k + " " + v + "\n"
	}

	// subvolumes <subvol1,subvol2..>
	if len(subvolumes) > 0 {
		out += "    subvolumes " + strings.Join(subvolumes, " ") + "\n"
	}

	// end volume
	out += "end-volume\n\n"
	return out, nil
}
