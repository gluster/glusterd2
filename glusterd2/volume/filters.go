package volume

const (
	// BlockHosted is plugin name for FilterBlockHostedVolumes
	BlockHosted = "block-hosted"
)

// Filter will receive a slice of *Volinfo and filters out the undesired one and return slice of desired one only
type Filter func([]*Volinfo) []*Volinfo

var filters = make(map[string]Filter)

// InstallFilter will register a custom Filter
func InstallFilter(name string, f Filter) {
	filters[name] = f
}

// ApplyFilters applies all registered filters passed in the args to a slice of *Volinfo
func ApplyFilters(volumes []*Volinfo, names ...string) []*Volinfo {
	for _, name := range names {
		if filter, found := filters[name]; found {
			volumes = filter(volumes)
		}
	}
	return volumes
}

// ApplyCustomFilters applies all custom filter to a slice of *Volinfo
func ApplyCustomFilters(volumes []*Volinfo, filters ...Filter) []*Volinfo {
	for _, filter := range filters {
		volumes = filter(volumes)
	}

	return volumes
}

// FilterBlockHostedVolumes filters out volume which are suitable for hosting block volume
func FilterBlockHostedVolumes(volumes []*Volinfo) []*Volinfo {
	var volInfos []*Volinfo
	for _, volume := range volumes {
		val, found := volume.Metadata[BlockHosting]
		if found && val == "yes" {
			volInfos = append(volInfos, volume)
		}
	}
	return volInfos
}

func init() {
	InstallFilter(BlockHosted, FilterBlockHostedVolumes)
}
