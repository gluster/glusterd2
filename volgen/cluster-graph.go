package volgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/volume"
)

const (
	clusterGraphType = "cluster.graph"
)

var (
	ErrClusterNoChild              = errors.New("cluster nodes cannot have children")
	ErrInvalidClusterGraphTemplate = errors.New("invalid cluster graph template")
	ErrIncorrectBricks             = errors.New("incorrect number of bricks given for volume")
)

func isClusterGraph(voltype string) bool {
	return voltype == clusterGraphType
}

func newClusterGraph(a qArgs) (*Node, error) {
	if len(a.t.Children) != 0 {
		return nil, ErrClusterNoChild
	}

	g, err := getTemplate(strings.ToLower(a.vol.Type.String()) + ".graph")
	if err != nil {
		return nil, err
	}

	n, err := processClusterGraph(g.root, a.vol)
	if err != nil {
		return nil, err
	}

	// At the end, we should have a single node in the returned list which we return
	// If not, we had an incorrect number of bricks for the volume
	if len(n) != 1 {
		return nil, ErrIncorrectBricks
	}

	return n[0], nil
}

func processClusterGraph(t *Node, vol *volume.Volinfo) ([]*Node, error) {
	// Cluster graphs need to be linear and cannot have branches
	// All xlators at a level in a cluster graph should be the same
	if len(t.Children) > 1 {
		return nil, ErrInvalidClusterGraphTemplate
	}

	var (
		siblings, descendents []*Node
		err                   error
	)

	if len(t.Children) == 1 {
		descendents, err = processClusterGraph(t.Children[0], vol)
		if err != nil {
			return nil, err
		}
	}

	// Special case for protocol/client
	if t.Voltype == "protocol/client" {
		return newClientNodes(vol)
	}

	sc := getChildCount(t.Voltype, vol)
	j, k := 0, 0
	var n *Node
	for _, d := range descendents {
		if j%sc == 0 {
			n = NewNode()
			n.Voltype = t.Voltype
			n.Id = fmt.Sprintf("%s-%s-%d", vol.Name, t.Id, k)
			if err := setOptions(n, vol); err != nil {
				_ = err
				//return nil, err
			}
			siblings = append(siblings, n)
			k++
		}
		n.Children = append(n.Children, d)
		j++
	}
	return siblings, nil
}

// Hardcoded for now. Need a way to avoid this
func getChildCount(t string, vol *volume.Volinfo) int {
	switch t {
	case "cluster/afr":
		fallthrough
	case "cluster/replicate":
		return vol.ReplicaCount
	case "cluster/dht":
		fallthrough
	case "cluster/distribute":
		return vol.DistCount
	default:
		return 0
	}
}

func newClientNodes(vol *volume.Volinfo) ([]*Node, error) {
	var ns []*Node

	for _, b := range vol.Bricks {
		n, err := newClientNode(vol, &b)
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}

	return ns, nil
}

func newClientNode(vol *volume.Volinfo, b *brick.Brickinfo) (*Node, error) {

	n := NewNode()
	n.Id = fmt.Sprintf("%s-client-%s", vol.Name, b.ID.String())
	n.Voltype = "protocol/client"

	if err := setOptions(n, vol); err != nil {
		return nil, err
	}

	return n, nil
}
