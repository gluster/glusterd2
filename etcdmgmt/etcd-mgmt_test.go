package etcdmgmt

import (
	"testing"

	"github.com/gluster/glusterd2/tests"
	heketitests "github.com/heketi/tests"
)

func TestStartETCD(t *testing.T) {
	etcdCmd, err := StartInitialEtcd()
	tests.Assert(t, err == nil)
	etcdCmd.Process.Kill()
}

func TestStartETCDWithInvalidExecName(t *testing.T) {
	// Mock the executable name such that it fails
	defer heketitests.Patch(&ExecName, "abc").Restore()
	_, err := StartInitialEtcd()
	tests.Assert(t, err != nil)
}
