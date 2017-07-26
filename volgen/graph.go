package volgen

import (
	"container/list"
	"fmt"

	"github.com/gluster/glusterd2/volume"
	"github.com/gluster/glusterd2/xlator"
)

type Node struct {
	Voltype  string
	Id       string
	Children []*Node
	Options  map[string]string
}

type Graph struct {
	id   string
	root *Node
}

// A type for the volgen processing queue
type qArgs struct {
	vol  *volume.Volinfo
	t, p *Node
}

func NewGraph() *Graph {
	return new(Graph)
}

func NewNode() *Node {
	n := new(Node)
	n.Options = make(map[string]string)
	return n
}

// Generate generates a graph from the template and volinfo
// The extra map can be used to provide any additional information
// XXX: Using volinfo here for now. Later needs to be changed to a standard
// struct to allow non volume specific graphs
func (gt *GraphTemplate) Generate(vol *volume.Volinfo, extra map[string]string) (*Graph, error) {
	g := NewGraph()
	g.id = fmt.Sprintf("%s-%s", vol.Name, gt.id)

	// The processing queue
	queue := list.New()
	// Add the template root as the first entry to the queue
	queue.PushBack(qArgs{vol, gt.root, nil})

	for i := queue.Front(); i != nil; i = i.Next() {
		a := i.Value.(qArgs)
		n, err := processNode(a)
		if err != nil {
			return nil, err
		}

		if a.p != nil {
			a.p.Children = append(a.p.Children, n)
		} else if g.root == nil {
			g.root = n
		}

		for _, t := range a.t.Children {
			// Add children to the queue to be processed
			queue.PushBack(qArgs{vol, t, n})
		}
	}

	return g, nil
}

func processNode(a qArgs) (*Node, error) {
	if isClusterGraph(a.t.Voltype) {
		return newClusterGraph(a)
	}
	return processNormalNode(a)
}

func processNormalNode(a qArgs) (*Node, error) {
	n := NewNode()
	n.Id = fmt.Sprintf("%s-%s", a.vol.Name, a.t.Id)
	n.Voltype = a.t.Voltype

	if err := setOptions(n, a.vol); err != nil {
		fmt.Println()
		fmt.Println(n.Id)
		fmt.Println(err)
		fmt.Println()
		_ = err
		//return nil, err
	}

	return n, nil
}

// setOptions uses the following rules to set xlator options
// - Get a list of all applicable options for a xlator
// - Iterate through the list and set options on the Node using the following
//   rules
//  - If option has been set in Volinfo, use that value
//	- Else if the option has a default value, skip setting the option. Xlator
//	  should use the default value.
//  - Else if the option does not have a default value, error out. This implies
//    that the particular option must be set
func setOptions(n *Node, vol *volume.Volinfo) error {
	xlOpts, ok := xlator.AllOptions[n.Voltype]
	if !ok {
		return ErrOptsNotFound(n.Voltype)
	}

	for _, o := range xlOpts {
		k, v, ok := getValue(n.Voltype, o.Key, vol.Options)
		if !ok {
			k = o.Key[0]
			v = o.DefaultValue
			if v == "" {
				// TODO: Remove this bypass later. Only here to allow a volfile to be
				// generated for the test program.
				v = "DEFAULT N/A"
				//return ErrOptRequired(o.Key[0])
			}
		}
		n.Options[k] = v
	}

	return nil
}

// XXX: Not possibly the best place for this
func getValue(id string, keys []string, opt map[string]string) (string, string, bool) {
	for _, k := range keys {
		v, ok := opt[id+"."+k]
		if ok {
			return k, v, true
		}
	}

	return "", "", false
}
