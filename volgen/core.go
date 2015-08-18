//Core file for graph and volfile generation

package volgen

import (
	"os"

	"github.com/kshlm/glusterd2/volume"
)

// GenerateVolfile function will do all task from graph generation to volfile generation
func GenerateVolfile(vinfo *volume.Volinfo) {

	graph := GenerateGraph(vinfo)

	f, err := os.Create("/tmp/client")
	if err != nil {
		panic(err)
	}
	defer closeFile(f)
	graph.DumpGraph(f)
}

func closeFile(f *os.File) {
	f.Close()
}
