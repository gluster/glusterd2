package volgen

// VolfileLevel is the level in which volfile need to be generated
type VolfileLevel uint16

const (
	// VolfileLevelBrick represents brick level volfile
	VolfileLevelBrick VolfileLevel = iota
	// VolfileLevelVolume represents volume level volfile
	VolfileLevelVolume
	// VolfileLevelCluster represents cluster level volfile
	VolfileLevelCluster
)

func (vl VolfileLevel) String() string {
	switch vl {
	case VolfileLevelCluster:
		return "cluster"
	case VolfileLevelVolume:
		return "volume"
	case VolfileLevelBrick:
		return "brick"
	default:
		return ""
	}
}

// Template represents Volfile template
type Template struct {
	// Name of template, this can be used to identify the template
	// while setting the template specific option like `<name>.<opt-name>`
	Name string `json:"name"`
	// Level represents the level in which volfiles will be generated
	// possible options are: cluster, volume and brick
	Level VolfileLevel `json:"level"`
	// Xlators represents the list of xlators to add in
	// the generated volfile
	Xlators []Xlator `json:"xlators"`
	// VolumeGraphXlators represents the list of xlators to add for
	// each volume. This list is only applicable for Cluster
	// level volfiles
	VolumeGraphXlators []Xlator `json:"volume-graph-xlators"`
	// SubvolGraphXlators represents the list of xlators to add for
	// each subvolume. This list is only applicable for cluster
	// and volume level volfiles
	SubvolGraphXlators []Xlator `json:"subvol-graph-xlators"`
	// BrickGraphXlators represents the list of xlators to add
	// for bricks. This list is only applicable for cluster and
	// volume volfiles. Brick level volfile will use Xlators itself
	// to define the list instead of BrickGraphXlators
	BrickGraphXlators []Xlator `json:"brick-graph-xlators"`
}

// Xlator represents Xlator in Volfile template
type Xlator struct {
	// NameTmpl can have template variables which will be
	// replaced while generating volfile. Replaced name will be
	// used as graph name in generated volfile
	NameTmpl string `json:"name-tmpl"`
	// Type represents xlator type. For example: "cluster/replicate"
	Type string `json:"type"`
	// TypeTmpl represents template, variables will be substituted
	// during volfile generation. Note: Variables should belong to the
	// level the xlator belongs to. For example, "{{ subvol.type }}"
	// can only be used if this xlator added to SubvolGraphXlators or
	// BrickGraphXlators.
	TypeTmpl string `json:"type-tmpl"`
	// OnlyLocalBricks can be used if this xlator will be added to
	// generated volfile if local bricks present in the current peer
	OnlyLocalBricks bool `json:"only-local-bricks"`
	// Disabled represents initial state of xlator in the generated
	// volfile. This will be flipped if the xlator is enabled in volinfo
	Disabled bool `json:"disabled"`
	// EnableByOption can be used if the xlator enabled state will be
	// decided based on option in the xlator graph in volfile with the
	// same name as of xlator name. For example: "option changelog on"
	EnableByOption bool `json:"enable-by-option"`
	// Options represents default options to include in the
	// generated volfile
	Options map[string]string `json:"options"`
	// IgnoreOptions represents list of options which should not be added in the
	// generated Volfile. For example, bitd and scrubd both uses same xlator, if
	// scrubber = on, then it becomes scrub daemon else it becomes bitd.
	IgnoreOptions []string `json:"ignore-options"`
}

// Templates represents collection of volfile templates
type Templates map[string]Template

const (
	// TemplateMetadataKey represents Volinfo Metadata key for customizing template names
	TemplateMetadataKey = "_template"
	// DefaultTemplateNamespace represents group of all default volfile templates
	DefaultTemplateNamespace = "default"
)

var namespaces = make(map[string]Templates)
