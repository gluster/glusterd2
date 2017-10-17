package volgen

import "errors"

var (
	// ErrClusterNoChild is returned when a `cluster.graph` node in a template has children
	ErrClusterNoChild = errors.New("cluster nodes cannot have children")
	// ErrInvalidClusterGraphTemplate is returned when a cluster graph template is not valid
	ErrInvalidClusterGraphTemplate = errors.New("invalid cluster graph template")
	// ErrIncorrectBricks is returned when not enough bricks are available when constructing the cluster graph
	ErrIncorrectBricks = errors.New("incorrect number of bricks given for volume")
)

// ErrOptsNotFound is returned when options for a xlator are not found in the options map
type ErrOptsNotFound string

func (e ErrOptsNotFound) Error() string {
	return "options not found for given xlator: " + string(e)
}
