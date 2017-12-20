package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	binDir   string
	functest bool
)

func TestMain(m *testing.M) {
	defBinDir, _ := filepath.Abs("../build")

	flag.BoolVar(&functest, "functest", false, "Run or skip functional test")
	flag.StringVar(&binDir, "bindir", defBinDir, "The directory containing glusterd2 binary")
	flag.Parse()

	if !functest {
		// Run only if -functest flag is passed to go test command
		// go test ./e2e -v -functest
		return
	}

	if os.Geteuid() != 0 {
		fmt.Println("Skipping functional tests (requires root)")
		return
	}

	// Cleanup leftovers from previous test runs. But don't cleanup after.
	os.RemoveAll("/tmp/gd2_func_test")

	v := m.Run()
	os.Exit(v)
}
