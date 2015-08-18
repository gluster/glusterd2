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
	graph := GenerateGraph(vinfo)

	//generate client volfile
	getClientFilePath(vinfo, &cpath)
	f, err := os.Create(cpath)
	if err != nil {
		panic(err)
	}
	defer closeFile(f)
	graph.DumpGraph(f)
}

func getVolumeDir(vinfo *volume.Volinfo, dir *string) {
	*dir = fmt.Sprintf("%s/vols/%s", workdir, vinfo.Name)
}

func getClientFilePath(vinfo *volume.Volinfo, path *string) {
	var cdir string
	getVolumeDir(vinfo, &cdir)

	// Create volume directory (/var/lib/glusterd/vols/<VOLNAME>)
	err := os.MkdirAll(cdir, 0666)
	if err != nil {
		panic(err)
	}
	*path = fmt.Sprintf("%s/trusted-%s.tcp-fuse.vol", cdir, vinfo.Name)
}

func closeFile(f *os.File) {
	f.Close()
}
