// Package volgen implements volume graph generation and volfile generation for GlusterD
package volgen

import (
	"fmt"
	"io"
)

// DumpGraph dumps a textual representation of the volume graph into `w`
func (graph Xlator_t) DumpGraph(w io.Writer) {

	for _, graph := range graph.Children {
		graph.DumpGraph(w)
	}

	fmt.Fprintf(w, "volume %s\n    type %s\n", graph.Name, graph.Type)

	for k, v := range graph.Options {
		fmt.Fprintf(w, "    options %v %v\n", k, v)
	}

	if graph.Children != nil {
		fmt.Fprintf(w, "    subvolumes")

		for _, graph := range graph.Children {
			fmt.Fprintf(w, " %v", graph.Name)
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "end-volume\n\n")
}
