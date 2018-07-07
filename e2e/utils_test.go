package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"
)

type testCluster struct {
	gds  []*gdProcess
	etcd *etcdProcess
}

// wrap takes a test function that requires the test type T and
// a test cluster instance and returns a function that only
// requires the test type (using the given test cluster).
func (tc *testCluster) wrap(
	f func(t *testing.T, c *testCluster)) func(*testing.T) {

	return func(t *testing.T) {
		f(t, tc)
	}
}

func setupCluster(configFiles ...string) (*testCluster, error) {

	tc := &testCluster{}

	cleanupRequired := true
	cleanup := func() {
		if cleanupRequired {
			for _, p := range tc.gds {
				p.Stop()
				p.EraseLocalStateDir()
			}
		}
	}
	defer cleanup()

	if externalEtcd {
		tc.etcd = &etcdProcess{
			DataDir: path.Join(baseLocalStateDir, "etcd/data"),
			LogPath: path.Join(baseLocalStateDir, "etcd/etcd.log"),
		}
		if err := os.MkdirAll(tc.etcd.DataDir, 0755); err != nil {
			return nil, err
		}
		err := tc.etcd.Spawn()
		if err != nil {
			return nil, err
		}
	}
	// exit the function early if no gd2 instances were requested
	if len(configFiles) == 0 {
		return tc, nil
	}

	for _, configFile := range configFiles {
		g, err := spawnGlusterd(configFile, true)
		if err != nil {
			return nil, err
		}
		tc.gds = append(tc.gds, g)
	}

	// restclient instance that will be used for peer operations
	client := initRestclient(tc.gds[0])

	// first gd2 instance spawned shall add other glusterd2 instances as its peers
	for i, gd := range tc.gds {
		if i == 0 {
			// do not add self
			continue
		}

		peerAddReq := api.PeerAddReq{
			Addresses: []string{gd.PeerAddress},
		}

		if _, err := client.PeerAdd(peerAddReq); err != nil {
			return nil, fmt.Errorf("setupCluster(): Peer add failed with error response %s",
				err.Error())
		}
	}

	// fail if the cluster hasn't been formed properly
	peers, err := client.Peers()
	if err != nil {
		return nil, err
	}

	if len(peers) != len(tc.gds) || len(peers) != len(configFiles) {
		return nil, fmt.Errorf("setupCluster() failed to create a cluster")
	}

	// do not run logic in cleanup() function that was deferred
	cleanupRequired = false

	return tc, nil
}

func teardownCluster(tc *testCluster) error {
	for _, gd := range tc.gds {
		gd.Stop()
	}
	if tc.etcd != nil {
		tc.etcd.Stop()
	}
	processes := []string{"glusterfs", "glusterfsd", "glustershd"}
	for _, p := range processes {
		exec.Command("killall", p).Run()
	}
	return nil
}

func initRestclient(gdp *gdProcess) *restclient.Client {
	secret := getAuth(gdp.LocalStateDir)
	return restclient.New("http://"+gdp.ClientAddress, "glustercli", secret, "", false)
}

func prepareLoopDevice(devname, loopnum, size string) error {
	err := exec.Command("fallocate", "-l", size, devname).Run()
	if err != nil {
		return err
	}

	err = exec.Command("mknod", "/dev/gluster_loop"+loopnum, "b", "7", loopnum).Run()
	if err != nil {
		return err
	}
	err = exec.Command("losetup", "/dev/gluster_loop"+loopnum, devname).Run()
	if err != nil {
		return err
	}
	return nil
}

func testlog(t *testing.T, msg string) {
	if t == nil {
		fmt.Println(msg)
		return
	}

	t.Log(msg)
}

func cleanupAllBrickMounts(t *testing.T) {
	// Unmount all Bricks in Working directory
	out, err := exec.Command("mount").Output()
	if err != nil {
		testlog(t, fmt.Sprintf("failed to list brick mounts: %s", err))
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// Identify Brick Mount
		if strings.Contains(line, baseLocalStateDir) {
			// Example: "/dev/mapper/gluster--vg--dev--gluster_loop2-brick_testvol--0--1 on \
			// /tmp/gd2_func_test/w1/mounts/testvol-0-1 type xfs (rw,noatime,seclabel, \
			// nouuid,attr2,inode64,logbsize=64k,sunit=128,swidth=2560,noquota
			parts := strings.Split(line, " ")
			if len(parts) < 3 {
				testlog(t, fmt.Sprintf("Unable to parse mount path: %s", line))
				continue
			}

			err = exec.Command("umount", parts[2]).Run()
			if err != nil {
				testlog(t, fmt.Sprintf("`umount %s` failed: %s", parts[2], err))
			}
		}
	}
}

func cleanupAllGlusterVgs(t *testing.T) {
	// List all Vgs and remove if it belongs to Gluster Testing
	out, err := exec.Command("vgs", "-o", "vg_name", "--no-headings").Output()
	if err == nil {
		vgs := strings.Split(string(out), "\n")
		for _, vg := range vgs {
			vg = strings.Trim(vg, " ")
			if strings.HasPrefix(vg, "vg-dev-gluster") {
				err = exec.Command("vgremove", "-f", vg).Run()
				if err != nil {
					testlog(t, fmt.Sprintf("`vgremove -f %s` failed: %s", vg, err))
				}
			}
		}
	}
}

func cleanupAllGlusterPvs(t *testing.T) {
	// Remove PV, detach and delete the loop device
	loopDevs, err := filepath.Glob("/dev/gluster_*")
	if err == nil {
		for _, loopDev := range loopDevs {
			err = exec.Command("pvremove", "-f", loopDev).Run()
			if err != nil {
				testlog(t, fmt.Sprintf("`pvremove -f %s` failed: %s", loopDev, err))
			}
			err = exec.Command("losetup", "-d", loopDev).Run()
			if err != nil {
				testlog(t, fmt.Sprintf("`losetup -d %s` failed: %s", loopDev, err))
			}
			err = os.Remove(loopDev)
			if err != nil {
				testlog(t, fmt.Sprintf("`rm %s` failed: %s", loopDev, err))
			}
		}
	}

}

func loopDevicesCleanup(t *testing.T) error {
	cleanupAllBrickMounts(t)
	cleanupAllGlusterVgs(t)
	cleanupAllGlusterPvs(t)

	// Cleanup device files
	devicefiles, err := filepath.Glob(baseLocalStateDir + "/*.img")
	if err == nil {
		for _, devicefile := range devicefiles {
			err := os.Remove(devicefile)
			if err != nil {
				testlog(t, fmt.Sprintf("`rm %s` failed: %s", devicefile, err))
			}
		}
	}
	return nil
}

func formatVolName(volName string) string {
	return strings.Replace(volName, "/", "-", 1)
}

func isProcessRunning(pidpath string) bool {
	content, err := ioutil.ReadFile(pidpath)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(content)))
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if err = process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

// testTempDir returns a temporary directory path that will exist
// on the system. This path is based on the name of the test and
// a unique final directory, determined by prefix.
// On encountering an error this function will panic.
func testTempDir(t *testing.T, prefix string) string {
	base := path.Join(baseLocalStateDir, t.Name())
	if err := os.MkdirAll(base, 0755); err != nil {
		panic(err)
	}
	d, err := ioutil.TempDir(base, prefix)
	if err != nil {
		panic(err)
	}
	return d
}

func getAuth(path string) string {
	filepath := path + "/auth"
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		s, err := ioutil.ReadFile(filepath)
		if err != nil {
			panic("unable to read auth file")
		}
		return string(s)
	}
	return ""
}
