package etcdmgmt

import (
	"fmt"
	"testing"

	"github.com/gluster/glusterd2/tests"
	heketitests "github.com/heketi/tests"
)

func TestStartETCDWithInvalidExecName(t *testing.T) {
	// Mock the executable name such that it fails
	defer heketitests.Patch(&ExecName, "abc").Restore()
	_, err := StartETCD()
	fmt.Println("error is: ", err)
	tests.Assert(t, err != nil)
}

func TestStartETCD(t *testing.T) {
	etcdCtx, err := StartETCD()
	tests.Assert(t, err == nil)
	etcdCtx.Kill()
}
