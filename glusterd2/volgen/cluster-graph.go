package volgen

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"
)

const (
	clusterGraphType = "cluster.graph"
)

func isClusterGraph(voltype string) bool {
	return voltype == clusterGraphType
}

func newClusterGraph(a qArgs) (*Node, error) {
	if len(a.t.Children) != 0 {
		return nil, ErrClusterNoChild
	}

	g, err := GetTemplate(strings.ToLower(a.vol.Type.String())+".graph", nil)
	if err != nil {
		return nil, err
	}

	n, err := processClusterGraph(g.root, a.vol, a.g.id, a.extra)
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

func processClusterGraph(t *Node, vol *volume.Volinfo, graph string, extra map[string]string) ([]*Node, error) {
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
		descendents, err = processClusterGraph(t.Children[0], vol, graph, extra)
		if err != nil {
			return nil, err
		}
	}

	// Special case for protocol/client
	if t.Voltype == "protocol/client" {
		return newClientNodes(vol, graph, extra)
	}

	sc := getChildCount(t.Voltype, vol)
	j, k := 0, 0
	var n *Node
	for _, d := range descendents {
		if j%sc == 0 {
			n = NewNode()
			n.Voltype = t.Voltype
			n.ID = fmt.Sprintf("%s-%s-%d", vol.Name, t.ID, k)
			if err := setOptions(n, graph, vol.Options, extra); err != nil {
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
		return vol.Subvols[0].ReplicaCount
	case "cluster/disperse":
		return vol.Subvols[0].DisperseCount
	case "cluster/dht":
		fallthrough
	case "cluster/distribute":
		return vol.DistCount
	default:
		return 0
	}
}

func newClientNodes(vol *volume.Volinfo, graph string, extra map[string]string) ([]*Node, error) {
	var ns []*Node

	for _, subvol := range vol.Subvols {
		for _, b := range subvol.Bricks {
			n, err := newClientNode(vol, &b, graph, extra)
			if err != nil {
				return nil, err
			}
			ns = append(ns, n)
		}
	}

	return ns, nil
}

func newClientNode(vol *volume.Volinfo, b *brick.Brickinfo, graph string, extra map[string]string) (*Node, error) {

	n := NewNode()
	n.ID = fmt.Sprintf("%s-client-%s", vol.Name, b.ID.String())
	n.Voltype = "protocol/client"
	extra = utils.MergeStringMaps(extra, b.StringMap())

	if err := setOptions(n, graph, vol.Options, extra); err != nil {
		return nil, err
	}

	return n, nil
}
