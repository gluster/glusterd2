package types

// Xlator represents a GlusterFS xlator
type Xlator struct {
	ID      string
	Options []*Option
}
