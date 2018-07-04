package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	toml "github.com/pelletier/go-toml"
)

type gdProcess struct {
	Cmd           *exec.Cmd
	ClientAddress string `toml:"clientaddress"`
	PeerAddress   string `toml:"peeraddress"`
	LocalStateDir string `toml:"localstatedir"`
	Rundir        string `toml:"rundir"`
	uuid          string
}

func (g *gdProcess) Stop() error {
	g.Cmd.Process.Signal(os.Interrupt) // try shutting down gracefully
	time.Sleep(500 * time.Millisecond)
	if g.IsRunning() {
		time.Sleep(1 * time.Second)
	} else {
		return nil
	}
	return g.Cmd.Process.Kill()
}

func (g *gdProcess) updateDirs() {
	g.Rundir = path.Clean(g.Rundir)
	if !path.IsAbs(g.Rundir) {
		g.Rundir = path.Join(baseLocalStateDir, g.Rundir)
	}
	g.LocalStateDir = path.Clean(g.LocalStateDir)
	if !path.IsAbs(g.LocalStateDir) {
		g.LocalStateDir = path.Join(baseLocalStateDir, g.LocalStateDir)
	}
}

func (g *gdProcess) EraseLocalStateDir() error {
	return os.RemoveAll(g.LocalStateDir)
}

func (g *gdProcess) IsRunning() bool {

	process, err := os.FindProcess(g.Cmd.Process.Pid)
	if err != nil {
		return false
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

func (g *gdProcess) PeerID() string {

	if g.uuid != "" {
		return g.uuid
	}

	// Endpoint doesn't matter here. All responses include a
	// X-Gluster-Peer-Id response header.
	endpoint := fmt.Sprintf("http://%s/version", g.ClientAddress)
	resp, err := http.Get(endpoint)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	g.uuid = resp.Header.Get("X-Gluster-Peer-Id")
	return g.uuid
}

func (g *gdProcess) IsRestServerUp() bool {

	endpoint := fmt.Sprintf("http://%s/v1/peers", g.ClientAddress)
	resp, err := http.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 5 {
		return false
	}

	return true
}

func spawnGlusterd(configFilePath string, cleanStart bool) (*gdProcess, error) {

	fContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	g := gdProcess{}
	if err = toml.Unmarshal(fContent, &g); err != nil {
		return nil, err
	}

	// The config files in e2e/config contain relative paths, convert them
	// to absolute paths.
	g.updateDirs()

	if cleanStart {
		g.EraseLocalStateDir() // cleanup leftovers from previous test
	}

	if err := os.MkdirAll(path.Join(g.LocalStateDir, "log"), os.ModeDir|os.ModePerm); err != nil {
		return nil, err
	}

	absConfigFilePath, err := filepath.Abs(configFilePath)
	if err != nil {
		return nil, err
	}
	g.Cmd = exec.Command(path.Join(binDir, "glusterd2"),
		"--config", absConfigFilePath,
		"--localstatedir", g.LocalStateDir,
		"--rundir", g.Rundir,
		"--logdir", path.Join(g.LocalStateDir, "log"),
		"--logfile", "glusterd2.log")

	if err := g.Cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		g.Cmd.Wait()
	}()

	retries := 4
	waitTime := 1500
	for i := 0; i < retries; i++ {
		// opposite of exponential backoff
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		if g.IsRestServerUp() {
			break
		}
		waitTime = waitTime / 2
	}

	if !g.IsRestServerUp() {
		return nil, errors.New("timeout: could not query rest server")
	}

	return &g, nil
}

func setupCluster(configFiles ...string) ([]*gdProcess, error) {

	var gds []*gdProcess

	cleanup := func() {
		for _, p := range gds {
			p.Stop()
			p.EraseLocalStateDir()
		}
	}

	for _, configFile := range configFiles {
		g, err := spawnGlusterd(configFile, true)
		if err != nil {
			cleanup()
			return nil, err
		}
		gds = append(gds, g)
	}

	// first gd2 that comes up shall add other nodes as its peers
	firstNode := gds[0].ClientAddress
	for i, gd := range gds {
		if i == 0 {
			continue
		}
		peerAddReq := api.PeerAddReq{
			Addresses: []string{gd.PeerAddress},
		}
		reqBody, errJSONMarshal := json.Marshal(peerAddReq)
		if errJSONMarshal != nil {
			cleanup()
			return nil, errJSONMarshal
		}

		resp, err := http.Post("http://"+firstNode+"/v1/peers", "application/json", strings.NewReader(string(reqBody)))
		if err != nil || resp.StatusCode != 201 {
			cleanup()
			return nil, err
		}
		resp.Body.Close()
	}

	return gds, nil
}

func teardownCluster(gds []*gdProcess) error {
	for _, gd := range gds {
		gd.Stop()
	}
	processes := []string{"glusterfs", "glusterfsd", "glustershd"}
	for _, p := range processes {
		exec.Command("killall", p).Run()
	}
	return nil
}

func initRestclient(clientAddress string) *restclient.Client {
	return restclient.New("http://"+clientAddress, "", "", "", false)
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
