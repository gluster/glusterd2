// Package volgen implements volume graph generation and volfile generation for GlusterD
package volgen

import (
	"fmt"
	"io"
)

// DumpGraph dumps a textual representation of the volume graph into `w`
func (x Xlator) DumpGraph(w io.Writer) {

	for _, xl := range x.Children {
		xl.DumpGraph(w)
	}

	fmt.Fprintf(w, "volume %s\n", x.Name)
	fmt.Fprintf(w, "   type %s", x.Type)

	for key, value := range x.Options {
		fmt.Fprintf(w, "\n   option %s %s", key, value)
	}

	if x.Children != nil {
		fmt.Fprintf(w, "\n   subvolumes")

		for _, xl := range x.Children {
			fmt.Fprintf(w, " %s", xl.Name)
		}
	}
	fmt.Fprintf(w, "\nend-volume\n\n")

}
