// xlator-checker checks if xlators have been analysed.
// It generates a YAML report of the xlators that probably need to be updated,
// and a list of xlators and the number of options updated

package main

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/glusterd2/xlator"

	"github.com/ghodss/yaml"
)

type xla struct {
	Name    string
	Options struct {
		Total, Updated int
	}
	ProbablyNeedsUpdate bool
}

var analysis struct {
	TotalXlators       int
	ProbablyNeedUpdate []string
	Xlators            []xla
	IgnoredXlators     []string
}

var ignoredXls = []string{"glusterd", "rot-13", "crypt", "tier", "bd", "fuse"}

func main() {
	if err := xlator.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load xlators: %s\n", err)
		os.Exit(1)
	}

	xls := xlator.Xlators()

	for _, xl := range xls {
		analyzeXl(xl)
	}

	out, err := yaml.Marshal(analysis)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error dumping analysis: %s", err.Error())
		os.Exit(1)
	}
	fmt.Print(string(out))
	fmt.Println()
}

// Check if the xlator has been updated
func analyzeXl(xl *xlator.Xlator) {
	var a xla

	if isIgnored(xl.ID) {
		analysis.IgnoredXlators = append(analysis.IgnoredXlators, xl.ID)
		return
	}

	a.Name = xl.ID
	a.Options.Total = len(xl.Options)

	// TODO: Check if Xlator has OpVersion set once it has been implemented
	for _, o := range xl.Options {
		// Check if OpVersion is set. This is the simplest check.
		for _, opv := range o.OpVersion {
			if opv != 0 {
				a.Options.Updated++
				break
			}
		}
	}

	if a.Options.Updated == 0 && a.Options.Total != 0 {
		a.ProbablyNeedsUpdate = true
		analysis.ProbablyNeedUpdate = append(analysis.ProbablyNeedUpdate, a.Name)
	}

	analysis.Xlators = append(analysis.Xlators, a)
}

func isIgnored(xl string) bool {
	for _, id := range ignoredXls {
		if xl == id {
			return true
		}
	}
	return false
}
