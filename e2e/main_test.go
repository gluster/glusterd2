package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	binDir            string
	baseLocalStateDir = "/tmp/gd2_func_test"
	functest          bool
	externalEtcd      = true
)

func TestMain(m *testing.M) {
	defBinDir, _ := filepath.Abs("../build")

	flag.BoolVar(&functest, "functest", false, "Run or skip functional test")
	flag.BoolVar(&externalEtcd, "external-etcd", true, "Run glusterd2 with an externally managed etcd")
	flag.StringVar(&binDir, "bindir", defBinDir, "The directory containing glusterd2 binary")
	flag.StringVar(&baseLocalStateDir, "basedir", baseLocalStateDir, "The base directory for test local state directories")
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

	// Cleanup leftover devices from previous test runs
	loopDevicesCleanup(nil)

	// Cleanup leftovers from previous test runs. But don't cleanup after.
	os.RemoveAll(baseLocalStateDir)

	v := m.Run()
	os.Exit(v)
}
