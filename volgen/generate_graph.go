/* Core file for volfile generation */

package volgen

import (
	"fmt"
	"os"
)

type trans struct {
	Name string
	Type string
}

func volgen_graph_add_as_root(graph *Xlator_t, vtype string) {
	switch Gtype {
	case "FUSE":
		graph.Name = Volname
		graph.Type = "debug/io-stats"

		graph.Options = make(map[string]string)

		// Add option to fuse graph
		graph.Options["count-fop-hits"] = "off"
		graph.Options["latency-measurement"] = "off"
	case "SERVER":
		gname := fmt.Sprintf("%s-server", Volname)
		graph.Name = gname
		graph.Type = "protocol/server"

		graph.Options = make(map[string]string)

		// Add option to server graph
		graph.Options["auth.addr./brr1.allow"] = "*"
		graph.Options["transport-type"] = "tcp"
	default:
		graph.Name = Daemon
		graph.Type = vtype
	}
}

func volgen_graph_add_client_link(cnode *Xlator_t, vtype string, name string) {
	node := new(Xlator_t)

	node.Options = make(map[string]string)

	node.Name = name
	node.Type = vtype

	hostname, _ := os.Hostname()

	// Add options to client subgraph
	node.Options["transport-type"] = "tcp"
	node.Options["remote-subvolume"] = "brick"
	node.Options["remote-host"] = hostname
	node.Options["ping-timeout"] = "42"

	cnode.Children = append(cnode.Children, *node)
}

func volgen_graph_build_client(vtype string, name string) *Xlator_t {
	cnode := new(Xlator_t)

	var i int

	switch Vtype {
	case "REPLICATE":
		for d := 0; d < Dcount; d++ {
			subnode := new(Xlator_t)
			for j := 1; j <= ReplicaCount; j++ {
				brick_id := fmt.Sprintf("%v-client-%v", Volname, i)
				volgen_graph_add_client_link(subnode, "protocol/client", brick_id)

				i++
			}
			sname := fmt.Sprintf("%s-replicate-%d", Volname, d)
			svtype := "cluster/replicate"
			subnode.Name = sname
			subnode.Type = svtype
			cnode.Children = append(cnode.Children, *subnode)
		}

		sname := fmt.Sprintf("%s-dht", Volname)
		svtype := "cluster/distribute"

		cnode.Name = sname
		cnode.Type = svtype
	default:
		// As of now if no volume type given then generate plane distribute volume graph
		for i := 0; i < Bcount; i++ {
			brick_id := fmt.Sprintf("%v-client-%v", Volname, i)
			volgen_graph_add_client_link(cnode, "protocol/client", brick_id)
		}

		cnode.Name = fmt.Sprintf("%s-dht", name)
		cnode.Type = vtype
	}

	return cnode
}

func volgen_graph_merge_client_with_root(Graph *Xlator_t, Craph *Xlator_t) {
	Graph.Children = append(Graph.Children, *Craph)
}

/* Adding options to translator*/
func volgen_graph_add_option(tgraph *Xlator_t) {
	tgraph.Options = make(map[string]string)

	switch tgraph.Type {
	case "storage/posix":
		tgraph.Options["update-link-count-parent"] = "on"
	case "features/trash":
		tgraph.Options["trash-internal-op"] = "off"
		tgraph.Options["trash-dir"] = ".trashcan"
	case "features/changetimerecorder":
		tgraph.Options["record-counters"] = "off"
		tgraph.Options["ctr-enabled"] = "off"
		tgraph.Options["record-entry"] = "on"
		tgraph.Options["ctr_inode_heal_expire_period"] = "300"
		tgraph.Options["ctr_link_consistency"] = "off"
		tgraph.Options["record-exit"] = "off"
		tgraph.Options["hot-brick"] = "off"
		tgraph.Options["db-type"] = "sqlite3"
	case "features/changelog":
		tgraph.Options["changelog-barrier-timeout"] = "120"
	case "features/upcall":
		tgraph.Options["cache-invalidation"] = "off"
	case "features/marker":
		tgraph.Options["inode-quota"] = "on"
		tgraph.Options["gsync-force-xtime"] = "off"
		tgraph.Options["xtime"] = "off"
	case "features/quota":
		tgraph.Options["deem-statfs"] = "on"
		tgraph.Options["timeout"] = "0"
		tgraph.Options["server-quota"] = "on"
	case "features/read-only":
		tgraph.Options["read-only"] = "off"
	default:
		tgraph.Options = nil
	}
}

func volgen_graph_link_parent_to_children(mgraph *Xlator_t, xdict []trans) *Xlator_t {
	var kgraph *Xlator_t
	for v, k := range xdict {
		tgraph := new(Xlator_t)
		tgraph.Name = fmt.Sprintf("%v-%v", Volname, k.Name)
		tgraph.Type = fmt.Sprintf("%v", k.Type)
		volgen_graph_add_option(tgraph)
		if v == 0 {
			mgraph.Children = append(mgraph.Children, *tgraph)
			(kgraph) = (&mgraph.Children[0])
			continue
		}
		kgraph.Children = append(kgraph.Children, *tgraph)
		kgraph = (&kgraph.Children[0])
	}

	return kgraph
}

func fuse_volgen_graph_build_xlator(cgraph *Xlator_t) *Xlator_t {
	var mgraph Xlator_t

	fdict := []trans{
		trans{Name: "open-behind", Type: "performance/open-behind"},
		trans{Name: "quick-read", Type: "performance/quick-read"},
		trans{Name: "io-cache", Type: "performance/io-cache"},
		trans{Name: "readdir-ahead", Type: "performance/readdir-ahead"},
		trans{Name: "read-ahead", Type: "performance/read-ahead"},
		trans{Name: "write-behind", Type: "performance/write-behind"},
	}

	mgraph.Name = fmt.Sprintf("%v-md-cache", Volname)
	mgraph.Type = fmt.Sprintf("performance/md-cache")

	/* Adding all above fdict[] translator to graph */
	fgraph := volgen_graph_link_parent_to_children(&mgraph, fdict)

	/* Appending all client graph as a child of write-behind translator*/
	fgraph.Children = append(fgraph.Children, *cgraph)

	return &mgraph
}

func server_volgen_graph_build_xlator() *Xlator_t {
	var mgraph Xlator_t

	sdict := []trans{
		trans{Name: "read-only", Type: "features/read-only"},
		trans{Name: "worm", Type: "features/worm"},
		trans{Name: "quota", Type: "features/quota"},
		trans{Name: "index", Type: "features/index"},
		trans{Name: "barrier", Type: "features/barrier"},
		trans{Name: "marker", Type: "features/marker"},
		trans{Name: "io-thread", Type: "performance/io-threads"},
		trans{Name: "upcall", Type: "features/upcall"},
		trans{Name: "locks", Type: "features/locks"},
		trans{Name: "access-control", Type: "features/access-control"},
		trans{Name: "bitrot-stub", Type: "features/bitrot-stub"},
		trans{Name: "changelog", Type: "features/changelog"},
		trans{Name: "changelogtimerecord", Type: "features/changetimerecorder"},
		trans{Name: "trash", Type: "features/trash"},
		trans{Name: "posix", Type: "storage/posix"},
	}

	mgraph.Name = fmt.Sprintf("brick")
	mgraph.Type = fmt.Sprintf("debug/io-stats")
	mgraph.Options = make(map[string]string)
	mgraph.Options["count-fop-hits"] = "off"
	mgraph.Options["latency-measurement"] = "off"

	/* Adding all above sdict[] translator to graph */
	volgen_graph_link_parent_to_children(&mgraph, sdict)

	return &mgraph
}

func volgen_graph_build_xlator(Cgraph *Xlator_t, gtype string) *Xlator_t {
	mgraph := new(Xlator_t)

	switch gtype {
	case "FUSE":
		mgraph = fuse_volgen_graph_build_xlator(Cgraph)
	case "SERVER":
		mgraph = server_volgen_graph_build_xlator()
	}
	return mgraph
}

func Generate_graph() *Xlator_t {
	Graph := new(Xlator_t)
	Cgraph := new(Xlator_t)
	Mgraph := new(Xlator_t)

	// Root of the graph
	vtype := fmt.Sprintf("features/%s", Daemon)
	volgen_graph_add_as_root(Graph, vtype)

	// Building client graph
	// To Do: call below function for total number of volume. As of now
	// Its only for single volume
	// Do not build client graph for server volfile.
	if Gtype != "SERVER" {
		vtype = fmt.Sprintf("cluster/distribute")
		Cgraph = volgen_graph_build_client(vtype, Volname)
	}

	// Build the translator graph which will be added bw client and root of
	// the graph other wise in case of server graph merge server graph
	// with rest of the graph
	Mgraph = volgen_graph_build_xlator(Cgraph, Gtype)

	// merge root of the graph with rest of the graph
	volgen_graph_merge_client_with_root(Graph, Mgraph)

	return Graph
}
