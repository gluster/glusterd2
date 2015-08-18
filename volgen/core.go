//Core file for graph and volfile generation

package volgen

import (
	"fmt"
	"os"

	"github.com/kshlm/glusterd2/volume"
)

const (
	workdir = "/var/lib/glusterd"
)

// GenerateVolfile function will do all task from graph generation to volfile generation
func GenerateVolfile(vinfo *volume.Volinfo) {
	var cpath string
	graph := GenerateGraph(vinfo, "CLIENT")

	//Generate client volfile
	getClientFilePath(vinfo, &cpath)
	f, err := os.Create(cpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	graph.DumpGraph(f)

	//Generate volfile for server
	for _, v := range vinfo.Bricks {
		sgraph := GenerateGraph(vinfo, "SERVER")

		//Generate volfile for server
		getServerFilePath(vinfo, &cpath, v)
		f, err := os.Create(cpath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sgraph.DumpGraph(f)
	}
}

func getVolumeDir(vinfo *volume.Volinfo, dir *string) {
	*dir = fmt.Sprintf("%s/vols/%s", workdir, vinfo.Name)
}

func getServerFilePath(vinfo *volume.Volinfo, path *string, brick string) {
	var vdir string
	getVolumeDir(vinfo, &vdir)

	hname, _ := os.Hostname()
	// Create volume directory (/var/lib/glusterd/vols/<VOLNAME>)
	err := os.MkdirAll(vdir, 0666)
	if err != nil {
		panic(err)
	}
	*path = fmt.Sprintf("%s/%s.%s.%s.vol", vdir, vinfo.Name, hname, brick)
}

func getClientFilePath(vinfo *volume.Volinfo, path *string) {
	var vdir string
	getVolumeDir(vinfo, &vdir)

	// Create volume directory (/var/lib/glusterd/vols/<VOLNAME>)
	err := os.MkdirAll(vdir, 0666)
	if err != nil {
		panic(err)
	}
	*path = fmt.Sprintf("%s/trusted-%s.tcp-fuse.vol", vdir, vinfo.Name)
}
