package volgen

import (
	"container/list"
	"fmt"
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/utils"
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
	g     *Graph
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
	g.id = gt.id

	extra = utils.MergeStringMaps(vol.StringMap(), extra)

	// The processing queue
	queue := list.New()
	// Add the template root as the first entry to the queue
	queue.PushBack(qArgs{vol, g, extra, gt.root, nil})

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
			queue.PushBack(qArgs{vol, g, extra, t, n})
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
	n.Voltype = a.t.Voltype

	// If template node ID is a varstring, do a varstring replacement and set it as the node ID.
	// Else, set node ID to "<volname>-<template node ID>"
	if isVarStr(a.t.ID) {
		id, err := varStrReplace(a.t.ID, a.extra)
		if err != nil {
			return nil, err
		}
		n.ID = id
	} else {
		n.ID = fmt.Sprintf("%s-%s", a.vol.Name, a.t.ID)
	}

	if err := setOptions(n, a.g.id, a.vol.Options, a.extra); err != nil {
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
//	- If the option has an explicit SetKey, use that as the key
//  - If option has been set in Volinfo, use that value
// 	- Else if the defaultvalue and the key are not varstrings, skip setting
// 		  the option.
// 	- If the key and value are varstring do varstring replacement
// 	- Set the key and value in the xlator options map
func setOptions(n *Node, graph string, opts, extra map[string]string) error {
	var (
		xl  *xlator.Xlator
		err error
	)

	xlid := path.Base(n.Voltype)
	xl, err = xlator.Find(xlid)
	if err != nil {
		return err
	}

	for _, o := range xl.Options {
		var (
			k, v string
			ok   bool
		)

		// If the option has an explicit SetKey, use it as the key
		if o.SetKey != "" {
			k = o.SetKey
			_, v, ok = getValue(graph, xlid, o.Key, opts)
		} else {
			k, v, ok = getValue(graph, xlid, o.Key, opts)
		}

		// If the option is not found in Volinfo, try to set to defaults if
		// available and required
		if !ok {
			// If there is no default value skip setting this option
			if o.DefaultValue == "" {
				continue
			}
			v = o.DefaultValue

			if k == "" {
				k = o.Key[0]
			}

			// If neither key nor value is a varstring, skip setting this option
			if !isVarStr(k) && !isVarStr(v) {
				continue
			}
		}

		// Do varsting replacements if required
		if isVarStr(k) {
			if k, err = varStrReplace(k, extra); err != nil {
				return err
			}
		}
		if isVarStr(v) {
			if v, err = varStrReplace(v, extra); err != nil {
				return err
			}
		}
		// Set the option
		n.Options[k] = v
	}

	return nil
}

// getValue returns value if found for provided graph.xlator.keys in the options map
// XXX: Not possibly the best place for this
func getValue(graph, xl string, keys []string, opts map[string]string) (string, string, bool) {
	graph = strings.TrimSuffix(graph, templateExt)

	for _, k := range keys {
		v, ok := opts[graph+"."+xl+"."+k]
		if ok {
			return k, v, true
		}
		v, ok = opts[xl+"."+k]
		if ok {
			return k, v, true
		}
	}

	return "", "", false
}
