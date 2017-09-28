package volgen

import (
	"container/list"
	"fmt"
	"path"

	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"
	"github.com/gluster/glusterd2/xlator"
)

// Node is an xlator in the GlusterFS volume graph
type Node struct {
	Voltype  string
	ID       string
	Children []*Node
	Options  map[string]string
}

// Graph is the GlusterFS volume graph
type Graph struct {
	id   string
	root *Node
}

// A type for the volgen processing queue
type qArgs struct {
	vol   *volume.Volinfo
	extra map[string]string
	t, p  *Node
}

// NewGraph returns an empty graph
func NewGraph() *Graph {
	return new(Graph)
}

// NewNode returns an empty graph node
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

	extra = utils.MergeStringMaps(vol.StringMap(), extra)

	// The processing queue
	queue := list.New()
	// Add the template root as the first entry to the queue
	queue.PushBack(qArgs{vol, extra, gt.root, nil})

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
			queue.PushBack(qArgs{vol, extra, t, n})
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
	n.ID = fmt.Sprintf("%s-%s", a.vol.Name, a.t.ID)
	n.Voltype = a.t.Voltype

	if err := setOptions(n, a.vol.Options, a.extra); err != nil {
		fmt.Println()
		fmt.Println(n.ID)
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
//	- Else if the option has a default value and is not a varstring, skip
//		setting the option. Xlator should use the default value.
// - If the value is a varstring do varstring replacement
func setOptions(n *Node, opts, extra map[string]string) error {
	var err error

	xl := path.Base(n.Voltype)
	xlOpts, ok := xlator.AllOptions[xl]
	if !ok {
		return ErrOptsNotFound(n.Voltype)
	}

	for _, o := range xlOpts {
		k, v, ok := getValue(xl, o.Key, opts)
		if !ok {
			if !isVarStr(o.DefaultValue) {
				continue
			}
			k = o.Key[0]
			v = o.DefaultValue
		}
		if isVarStr(v) {
			if v, err = varStrReplace(v, extra); err != nil {
				return err
			}
		}
		n.Options[k] = v
	}

	return nil
}

// getValue returns value if found for provided xlator.keys in the options map
// XXX: Not possibly the best place for this
func getValue(xl string, keys []string, opts map[string]string) (string, string, bool) {
	for _, k := range keys {
		v, ok := opts[xl+"."+k]
		if ok {
			return k, v, true
		}
	}

	return "", "", false
}
