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
	Name               string       `json:"name"`
	Level              VolfileLevel `json:"level"`
	Xlators            []Xlator     `json:"xlators"`
	VolumeGraphXlators []Xlator     `json:"volume-graph-xlators"`
	SubvolGraphXlators []Xlator     `json:"subvol-graph-xlators"`
	BrickGraphXlators  []Xlator     `json:"brick-graph-xlators"`
}

// Xlator represents Xlator in Volfile template
type Xlator struct {
	NameTmpl        string            `json:"name-tmpl"`
	Type            string            `json:"type"`
	TypeTmpl        string            `json:"type-tmpl"`
	OnlyLocalBricks bool              `json:"only-local-bricks"`
	Disabled        bool              `json:"disabled"`
	EnableByOption  bool              `json:"enable-by-option"`
	Options         map[string]string `json:"options"`
	IgnoreOptions   []string          `json:"ignore-options"`
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
