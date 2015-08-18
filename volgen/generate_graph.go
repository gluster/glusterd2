/* Core file for volfile generation */

package volgen

import (
	"fmt"
	"os"

	"github.com/kshlm/glusterd2/volume"
)

type trans struct {
	Name string
	Type string
}

func graphAddAsRoot(graph *Xlator, vinfo *volume.Volinfo, gtype string) {
	switch gtype {
	case "FUSE":
		graph.Name = vinfo.Name
		graph.Type = "debug/io-stats"

		graph.Options = make(map[string]string)

		// Add option to fuse graph
		graph.Options["count-fop-hits"] = "off"
		graph.Options["latency-measurement"] = "off"
	case "SERVER":
		gname := fmt.Sprintf("%s-server", vinfo.Name)
		graph.Name = gname
		graph.Type = "protocol/server"

		graph.Options = make(map[string]string)

		// Add option to server graph
		graph.Options["auth.addr./brr1.allow"] = "*"
		graph.Options["transport-type"] = "tcp"
	default:
		os.Exit(2)
	}
}

func addGraphClientLink(cnode *Xlator, vtype string, name string, brick string) {
	node := new(Xlator)

	node.Options = make(map[string]string)

	node.Name = name
	node.Type = vtype

	hostname, _ := os.Hostname()

	// Add options to client subgraph
	node.Options["transport-type"] = "tcp"
	node.Options["remote-subvolume"] = brick
	node.Options["remote-host"] = hostname
	node.Options["ping-timeout"] = "42"

	cnode.Children = append(cnode.Children, *node)
}

func graphBuildClient(vinfo *volume.Volinfo) *Xlator {
	cnode := new(Xlator)

	var i int

	switch vinfo.Type {
	case 1:
		lbrick := len(vinfo.Bricks)
		Dcount := lbrick / int(vinfo.ReplicaCount)

		for d := 0; d < Dcount; d++ {
			subnode := new(Xlator)
			for j := 1; j <= int(vinfo.ReplicaCount); j++ {
				name := fmt.Sprintf("%v-client-%v", vinfo.Name, i)
				addGraphClientLink(subnode, "protocol/client", name, vinfo.Bricks[i])

				i++
			}
			sname := fmt.Sprintf("%s-replicate-%d", vinfo.Name, d)
			svtype := "cluster/replicate"
			subnode.Name = sname
			subnode.Type = svtype
			cnode.Children = append(cnode.Children, *subnode)
		}

		sname := fmt.Sprintf("%s-dht", vinfo.Name)
		svtype := "cluster/distribute"

		cnode.Name = sname
		cnode.Type = svtype
	default:
		// As of now if no volume type given then generate plane distribute volume graph
		for i, brick := range vinfo.Bricks {
			name := fmt.Sprintf("%v-client-%v", vinfo.Name, i)
			addGraphClientLink(cnode, "protocol/client", name, brick)
			i++
		}

		cnode.Name = fmt.Sprintf("%s-dht", vinfo.Name)
		cnode.Type = "cluster/distribute"
	}

	return cnode
}

func mergeClientWithRoot(Graph *Xlator, Craph *Xlator) {
	Graph.Children = append(Graph.Children, *Craph)
}

/* Adding options to translator*/
func addOption(tgraph *Xlator) {
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

func linkParentToChildren(vinfo *volume.Volinfo, mgraph *Xlator, xdict []trans) *Xlator {
	var kgraph *Xlator
	for v, k := range xdict {
		tgraph := new(Xlator)
		tgraph.Name = fmt.Sprintf("%v-%v", vinfo.Name, k.Name)
		tgraph.Type = fmt.Sprintf("%v", k.Type)
		addOption(tgraph)
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

func fuseBuildXlator(vinfo *volume.Volinfo, cgraph *Xlator) *Xlator {
	var mgraph Xlator

	fdict := []trans{
		trans{Name: "open-behind", Type: "performance/open-behind"},
		trans{Name: "quick-read", Type: "performance/quick-read"},
		trans{Name: "io-cache", Type: "performance/io-cache"},
		trans{Name: "readdir-ahead", Type: "performance/readdir-ahead"},
		trans{Name: "read-ahead", Type: "performance/read-ahead"},
		trans{Name: "write-behind", Type: "performance/write-behind"},
	}

	mgraph.Name = fmt.Sprintf("%v-md-cache", vinfo.Name)
	mgraph.Type = fmt.Sprintf("performance/md-cache")

	/* Adding all above fdict[] translator to graph */
	fgraph := linkParentToChildren(vinfo, &mgraph, fdict)

	/* Appending all client graph as a child of write-behind translator*/
	fgraph.Children = append(fgraph.Children, *cgraph)

	return &mgraph
}

func serverBuildXlator(vinfo *volume.Volinfo) *Xlator {
	var mgraph Xlator

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
	linkParentToChildren(vinfo, &mgraph, sdict)

	return &mgraph
}

func buildXlator(vinfo *volume.Volinfo, Cgraph *Xlator, gtype string) *Xlator {
	mgraph := new(Xlator)

	switch gtype {
	case "FUSE":
		mgraph = fuseBuildXlator(vinfo, Cgraph)
	case "SERVER":
		mgraph = serverBuildXlator(vinfo)
	}
	return mgraph
}

//GenerateGraph function will do all task for graph generation
func GenerateGraph(vinfo *volume.Volinfo) *Xlator {
	Graph := new(Xlator)
	Cgraph := new(Xlator)
	Mgraph := new(Xlator)

	// Root of the graph
	//vtype := fmt.Sprintf("features/%s", Daemon)
	graphAddAsRoot(Graph, vinfo, "FUSE")

	// Building client graph
	// To Do: call below function for total number of volume. As of now
	// Its only for single volume
	// Do not build client graph for server volfile.
	//if Gtype != "SERVER" {
	Cgraph = graphBuildClient(vinfo)
	//}

	// Build the translator graph which will be added bw client and root of
	// the graph other wise in case of server graph merge server graph
	// with rest of the graph
	Mgraph = buildXlator(vinfo, Cgraph, "FUSE")

	// merge root of the graph with rest of the graph
	mergeClientWithRoot(Graph, Mgraph)

	return Graph
}
