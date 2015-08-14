package volgen

// Xlator is a node in the GlusterFS volume graph
type Xlator struct {
	Name     string
	Type     string
	Options  map[string]string
	Children []Xlator
}
