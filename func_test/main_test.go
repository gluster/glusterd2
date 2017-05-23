package func_test

import (
	"flag"
	"os"
	"testing"
)

var binDir string
var functest bool

func TestMain(m *testing.M) {

	flag.BoolVar(&functest, "functest", false,
		"Run or skip functional test")

	flag.StringVar(&binDir, "bindir", "../build",
		"The directory containing glusterd2 binary")

	flag.Parse()

	if !functest {
		// Run only if -functest flag is passed to go test command
		// go test -tags 'novirt noaugeas' ./func_test -v -functest
		return
	}

	v := m.Run()
	os.Exit(v)
}
