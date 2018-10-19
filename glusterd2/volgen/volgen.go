package volgen

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

// Entry represents one Xlator entry in Volfile
type Entry struct {
	Name       string
	NamePrefix string
	XlatorData Xlator
	SubEntries []Entry
	VarStrData map[string]string
	// IgnoreOptions represents list of options which should not be added in the
	// generated Volfile. For example, bitd and scrubd both uses same xlator, if
	// scrubber = on, then it becomes scrub daemon else it becomes bitd.
}

// Volfile represents Gluster Volfile
type Volfile struct {
	Name      string
	RootEntry Entry
}

// NewVolfile initializes Entry structure
func NewVolfile(name string) *Volfile {
	return &Volfile{Name: name, RootEntry: Entry{}}
}

// Add adds sub entry
func (e *Entry) Add(xl Xlator, data map[string]string) *Entry {
	e.SubEntries = append(e.SubEntries, Entry{XlatorData: xl, VarStrData: data})

	// Return the last element's pointer, useful if sub elements to be added to the newly added element
	return &e.SubEntries[len(e.SubEntries)-1]
}

// SetNamePrefix sets name prefix
func (e *Entry) SetNamePrefix(name string) *Entry {
	e.NamePrefix = name
	return e
}

// Generate generates Volfile content
func (v *Volfile) Generate() (string, error) {
	return v.RootEntry.Generate()
}

func setNameAndType(entry *Entry) error {
	if entry.XlatorData.Type == "" && entry.XlatorData.TypeTmpl != "" {
		ty, err := varStrReplace(entry.XlatorData.TypeTmpl, entry.VarStrData)
		if err != nil {
			return err
		}
		entry.XlatorData.Type = ty
	}
	if entry.XlatorData.NameTmpl != "" {
		name, err := varStrReplace(entry.XlatorData.NameTmpl, entry.VarStrData)
		if err != nil {
			return err
		}
		entry.Name = name
	}
	if entry.Name == "" {
		// If Xlator name template is not specified, construct the xlator
		// graph name as <volume-name>-<xlator-suffix>
		prefix := entry.NamePrefix
		if prefix != "" {
			prefix = prefix + "-"
		}

		if entry.NamePrefix == "" {
			volname, exists := entry.VarStrData["volume.name"]
			if exists {
				prefix = volname + "-"
			}
		}
		entry.Name = prefix + entry.XlatorData.suffix()
	}
	return nil
}

// Generate generates the Volfile content
func (e *Entry) Generate() (string, error) {
	out := ""

	subvolumes := []string{}
	for _, entry := range e.SubEntries {
		err := setNameAndType(&entry)
		if err != nil {
			return "", err
		}
		out1, err := entry.Generate()
		if err != nil {
			return "", err
		}
		out += out1
		subvolumes = append(subvolumes, entry.Name)
	}

	if e.XlatorData.Type == "" {
		return out, nil
	}

	err := setNameAndType(e)
	if err != nil {
		return "", err
	}

	// volume <name>
	out += "volume " + e.Name + "\n"

	// type <type>
	out += "    type " + e.XlatorData.Type + "\n"

	// Remove if any options needs to be ignored
	for k, v := range e.XlatorData.Options {
		ignore := false
		for _, ignoreOpt := range e.XlatorData.IgnoreOptions {
			if ignoreOpt == k {
				ignore = true
				break
			}
		}
		if !ignore {
			if isVarStr(k) {
				k, err = varStrReplace(k, e.VarStrData)
				if err != nil {
					return "", err
				}
			}

			if isVarStr(v) {
				v, err = varStrReplace(v, e.VarStrData)
				if err != nil {
					return "", err
				}
			}

			if strings.TrimSpace(v) != "" {
				// option <key> <value>
				out += "    option " + k + " " + v + "\n"
			}
		}
	}

	// subvolumes <subvol1,subvol2..>
	if len(subvolumes) > 0 {
		out += "    subvolumes " + strings.Join(subvolumes, " ") + "\n"
	}

	// end volume
	out += "end-volume\n\n"
	return out, nil
}

// BrickLevelVolfile generates brick level volfile
func BrickLevelVolfile(tmpl *Template, volinfo *volume.Volinfo, peerid string, brickpath string) (string, error) {
	extraStringMaps := getExtraStringMaps(volinfo)
	varStrData := utils.MergeStringMaps(volinfo.StringMap(), extraStringMaps.StringMap)
	arbiterBrick := false

VolinfoLoop:
	for sidx, sv := range volinfo.Subvols {
		for bidx, b := range sv.Bricks {
			if peerid == b.PeerID.String() && brickpath == b.Path {
				if b.Type == brick.Arbiter {
					arbiterBrick = true
				}

				// Merge all string maps related to bricks
				varStrData = utils.MergeStringMaps(
					varStrData,
					sv.StringMap(),
					extraStringMaps.Subvols[sidx].StringMap,
					b.StringMap(),
					extraStringMaps.Subvols[sidx].Bricks[bidx].StringMap,
				)
				break VolinfoLoop
			}
		}
	}

	// Set Arbiter option if it is arbiter brick
	if arbiterBrick {
		volinfo.Options["brick.features/arbiter"] = "on"
	}

	// Xlators list from template
	xlators, err := tmpl.EnabledXlators(volinfo)
	if err != nil {
		return "", err
	}

	volfile := NewVolfile(tmpl.Name)
	entry := &volfile.RootEntry
	for _, xl := range xlators {
		entry = entry.Add(xl, varStrData)
	}

	return volfile.Generate()
}

func volumegraph(tmpl *Template, volinfo volume.Volinfo, entry *Entry, varStrData *map[string]string, extraStringMaps *stringMapVolume) error {
	numSubvols := len(volinfo.Subvols)
	// thin arbiter support, if thin arbiter is set then add virtual brick
	// to each sub volume, so that resulting volfile
	// will include that details
	thinarbiter, exists := volinfo.Options[thinArbiterOptionName]
	remotePort := thinArbiterDefaultPort

	if exists && thinarbiter != "" {
		taParts := strings.Split(thinarbiter, ":")
		if len(taParts) != 2 && len(taParts) != 3 {
			return errors.New("invalid thin arbiter brick details")
		}

		if len(taParts) >= 3 {
			remotePort = taParts[2]
		}

		// Slices are sent as reference, updating subvols directly
		// by index will update the global volinfo
		// volinfo.Subvols[sidx].Bricks = append(..
		// Copy the subvols list before updating
		subvols := volinfo.Subvols
		volinfo.Subvols = []volume.Subvol{}
		for sidx, sv := range subvols {
			volinfo.Subvols = append(volinfo.Subvols, sv)
			// Add extra virtual brick entry
			volinfo.Subvols[sidx].Bricks = append(
				volinfo.Subvols[sidx].Bricks,
				brick.Brickinfo{
					ID:         uuid.NewRandom(),
					Hostname:   taParts[0],
					Path:       taParts[1],
					VolumeName: volinfo.Name,
					VolumeID:   volinfo.ID,
					Type:       brick.ThinArbiter,
				},
			)
		}
		// Recreate extraStringMaps after adding thin arbiter virtual brick
		*extraStringMaps = getExtraStringMaps(&volinfo)
	}

	// Subvol Xlators list and Brick Xlators
	for sidx, sv := range volinfo.Subvols {
		subvolXlators, err := tmpl.EnabledSubvolGraphXlators(&volinfo, &sv)
		if err != nil {
			return err
		}

		numberOfLocalBricks := 0
		for _, b := range sv.Bricks {
			if b.PeerID.String() == gdctx.MyUUID.String() {
				numberOfLocalBricks++
			}
		}

		// Special handling: If subvol type is Distribute
		// and number of subvols is 1 then do not include
		// cluster/distribute graph again. Directly assign
		// brick entries to main cluster/distribute itself
		sentry := entry
		if sv.Type != volume.SubvolDistribute || (sv.Type == volume.SubvolDistribute && numSubvols > 1) {
			for _, sxl := range subvolXlators {
				if !sxl.OnlyLocalBricks || (sxl.OnlyLocalBricks && numberOfLocalBricks > 0) {
					sentry = sentry.Add(sxl, utils.MergeStringMaps(
						*varStrData,
						sv.StringMap(),
						extraStringMaps.Subvols[sidx].StringMap,
					)).SetNamePrefix(sv.Name)
				}
			}
		}

		if len(subvolXlators) == 0 {
			continue
		}

		for bidx, b := range sv.Bricks {
			brickXlators, err := tmpl.EnabledBrickGraphXlators(&volinfo, &sv, &b)
			if err != nil {
				return err
			}
			bentry := sentry
			for _, bxl := range brickXlators {
				if bxl.OnlyLocalBricks && b.PeerID.String() != gdctx.MyUUID.String() {
					continue
				}

				bopts := utils.MergeStringMaps(
					*varStrData,
					sv.StringMap(),
					extraStringMaps.Subvols[sidx].StringMap,
					b.StringMap(),
					extraStringMaps.Subvols[sidx].Bricks[bidx].StringMap,
				)

				// Add remote port if it is thin arbiter brick
				if b.Type == brick.ThinArbiter {
					bopts = utils.MergeStringMaps(
						bopts,
						map[string]string{"remote-port": remotePort},
					)
				}
				bentry = bentry.Add(bxl, bopts).SetNamePrefix(sv.Name + "-" + strconv.Itoa(bidx))
			}
		}
	}
	return nil
}

// VolumeLevelVolfile generates volume level volfile
func VolumeLevelVolfile(tmpl *Template, volinfo *volume.Volinfo) (string, error) {
	// Xlators list from template
	xlators, err := tmpl.EnabledXlators(volinfo)
	if err != nil {
		return "", err
	}
	extraStringMaps := getExtraStringMaps(volinfo)
	varStrData := utils.MergeStringMaps(volinfo.StringMap(), extraStringMaps.StringMap)

	volfile := NewVolfile(tmpl.Name)
	entry := &volfile.RootEntry

	// Global Xlators list
	for _, xl := range xlators {
		entry = entry.Add(xl, varStrData)
	}
	err = volumegraph(tmpl, *volinfo, entry, &varStrData, &extraStringMaps)
	if err != nil {
		return "", err
	}

	return volfile.Generate()
}

// ClusterLevelVolfile generates cluster level volfile
func ClusterLevelVolfile(tmpl *Template, clusterinfo []*volume.Volinfo) (string, error) {
	// Xlators list from template
	xlators, err := tmpl.EnabledXlators(nil)
	if err != nil {
		return "", err
	}
	volfile := NewVolfile(tmpl.Name)
	entry := &volfile.RootEntry

	// Global Xlators list
	for _, xl := range xlators {
		entry = entry.Add(xl, nil)
	}

	for _, volinfo := range clusterinfo {
		// Include only if Volume is started state
		if volinfo.State != volume.VolStarted {
			continue
		}

		extraStringMaps := getExtraStringMaps(volinfo)
		varStrData := utils.MergeStringMaps(volinfo.StringMap(), extraStringMaps.StringMap)
		volumeXlators, err := tmpl.EnabledVolumeGraphXlators(volinfo)
		if err != nil {
			return "", err
		}

		ventry := entry
		for _, xl := range volumeXlators {
			ventry = ventry.Add(xl, varStrData)
		}

		// If atleast one Volume Graph Xlator specified and zero enabled xlators
		if len(tmpl.VolumeGraphXlators) > 0 && len(volumeXlators) == 0 {
			continue
		}

		err = volumegraph(tmpl, *volinfo, ventry, &varStrData, &extraStringMaps)
		if err != nil {
			return "", err
		}
	}
	return volfile.Generate()
}
