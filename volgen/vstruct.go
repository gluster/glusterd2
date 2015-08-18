// Xlator is a node in the GlusterFS volume graph

package volgen

type Xlator_t struct {
	Name     string
	Type     string
	Options  map[string]string
	Children []Xlator_t
}
