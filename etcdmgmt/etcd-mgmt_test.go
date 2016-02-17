package etcdmgmt

import (
	"os"
	"os/exec"
	"testing"

	"github.com/gluster/glusterd2/tests"
	heketitests "github.com/heketi/tests"
)

func TestStartETCDWithInvalidExecName(t *testing.T) {
	// Mock the executable name such that it fails
	defer heketitests.Patch(&ExecName, "abc").Restore()
	_, err := StartETCD()
	tests.Assert(t, err != nil)
}

func TestStartETCD(t *testing.T) {
	etcdCtx, err := StartETCD()
	tests.Assert(t, err == nil)
	etcdCtx.Kill()
}

func TestWriteETCDPidFile(t *testing.T) {
	cmd := exec.Command("etcd")
	_ = cmd.Start()
	tests.Assert(t, writeETCDPidFile(cmd.Process.Pid) == nil)
	os.Remove(etcdPidFile)

	// change etcdPidFile location such that its an invalid path and
	// writeETCDPidFile should fail
	defer heketitests.Patch(&etcdPidFile, "/a/b/c/d/etcd.pid").Restore()
	tests.Assert(t, writeETCDPidFile(cmd.Process.Pid) != nil)
	cmd.Process.Kill()
}

func TestIsETCDStartNeeded(t *testing.T) {
	// check once etcd process is running isETCDStartNeeded returns false
	os.Remove(etcdPidFile)
	cmd := exec.Command("etcd")
	err := cmd.Start()
	tests.Assert(t, err == nil)
	err = writeETCDPidFile(cmd.Process.Pid)
	tests.Assert(t, err == nil)
	start, _ := isETCDStartNeeded()
	tests.Assert(t, start == false)

	//check once etcd process is killed isETCDStartNeeded returns true
	var pid int
	oldPid := cmd.Process.Pid
	err = cmd.Process.Kill()
	tests.Assert(t, err == nil)
	cmd.Wait()
	start, pid = isETCDStartNeeded()
	tests.Assert(t, oldPid == pid)
	tests.Assert(t, start == true)

	// check if the pid file is missing then isETCDStartNeeded returns true
	os.Remove(etcdPidFile)
	start, _ = isETCDStartNeeded()
	tests.Assert(t, start == true)
	cmd.Process.Kill()
}
