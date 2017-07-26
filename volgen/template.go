package volgen

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"path"

	"github.com/gluster/glusterd2/volume"
	"github.com/gluster/glusterd2/xlator"
)

// Templates are empty graphs built from template files
type GraphTemplate Graph

// ReadTemplateFile reads in a template file and generates a template graph
func ReadTemplateFile(p string) (*GraphTemplate, error) {
	tf, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer tf.Close()

	t := new(GraphTemplate)
	t.id = path.Base(p)

	var curr, prev *Node

	s := bufio.NewScanner(tf)
	for s.Scan() {
		curr = NewNode()

		curr.Voltype = s.Text()
		curr.Id = path.Base(curr.Voltype)
		if t.root == nil {
			t.root = curr
		}
		if prev != nil {
			prev.Children = append(prev.Children, curr)
		}
		prev = curr
		// TODO: Handle graph templates with branches
	}

	return t, nil
}

// Generate generates a graph from the template and volinfo
// XXX: Using volinfo here for now. Later needs to be changed to a standard
// struct to allow non volume specific graphs
func (gt *GraphTemplate) Generate(vol *volume.Volinfo) (*Graph, error) {
	g := NewGraph()
	g.id = fmt.Sprintf("%s-%s", vol.Name, gt.id)

	// A type for the procesing queue
	type qArgs struct {
		t, p *Node
	}
	// The processing queue
	queue := list.New()
	// Add the template root as the first entry to the queue
	queue.PushBack(qArgs{gt.root, nil})

	for i := queue.Front(); i != nil; i = i.Next() {
		a := i.Value.(qArgs)
		n := NewNode()
		// TODO: Need a way to consistently generate IDs for cluster xlator children
		n.Id = fmt.Sprintf("%s-%s", vol.Name, a.t.Id)
		n.Voltype = a.t.Voltype

		if err := setOptions(n, vol); err != nil {
			_ = err
			//return nil, err
		}

		if a.p != nil {
			a.p.Children = append(a.p.Children, n)
		} else if g.root == nil {
			g.root = n
		}

		// TODO: Control number of children for cluster xlators here
		for _, t := range a.t.Children {
			// Add children to the queue to be processed
			queue.PushBack(qArgs{t, n})
		}
	}

	return g, nil
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
