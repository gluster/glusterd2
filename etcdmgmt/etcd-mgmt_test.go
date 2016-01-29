package etcdmgmt

import (
	"testing"

	"github.com/gluster/glusterd2/tests"
	heketitests "github.com/heketi/tests"
)

func TestStartETCD(t *testing.T) {
	etcdCtx, err := StartETCD()
	tests.Assert(t, err == nil)
	etcdCtx.Kill()
}

func TestStartETCDWithInvalidExecName(t *testing.T) {
	// Mock the executable name such that it fails
	defer heketitests.Patch(&ExecName, "abc").Restore()
	_, err := StartETCD()
	tests.Assert(t, err != nil)
}
