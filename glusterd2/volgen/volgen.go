package volgen

import (
	"errors"
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/pborman/uuid"
)

// Entry represents one Xlator entry in Volfile
type Entry struct {
	Name     string
	Type     string
	VolumeID uuid.UUID
	BrickID  uuid.UUID
	// ExtraOptions represents additional options or override options
	// which are not automatically detected from xlators so or stored options
	// For example, Quotad uses "option <volname>.volume-id = <volname>"
	ExtraOptions map[string]string
	// ExtraData represents additional data which will be used for replacing
	// template variables used in option name or value. For example, gfproxy client
	// uses protocol/client at volume level, where brick info is not available,
	// so option remote-subvolume = {{brick.path}} will not get replaced.
	// In Glusterd1 generated volfile, this "brick.path" is filled with "gfproxyd-<volname>"
	// option remote-subvolume gfproxyd-<volname>
	ExtraData  map[string]string
	SubEntries []Entry
	// IgnoreOptions represents list of options which should not be added in the
	// generated Volfile. For example, bitd and scrubd both uses same xlator, if
	// scrubber = on, then it becomes scrub daemon else it becomes bitd.
	IgnoreOptions map[string]bool
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

// SetIgnoreOptions sets list of options not required to add in generated
// volfile
func (e *Entry) SetIgnoreOptions(opts []string) *Entry {
	e.IgnoreOptions = make(map[string]bool)
	for _, opt := range opts {
		e.IgnoreOptions[opt] = true
	}
	return e
}

// Generate generates Volfile content
func (v *Volfile) Generate(graph string, extra *map[string]extrainfo) (string, error) {
	if v.Name == "" || v.FileName == "" {
		return "", errors.New("incomplete details")
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
		if _, ok := e.IgnoreOptions[k]; ok {
			continue
		}
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
