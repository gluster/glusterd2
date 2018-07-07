package e2e

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	toml "github.com/pelletier/go-toml"
)

type testProcess struct {
	Cmd *exec.Cmd
}

// IsRunning will return true if the process is currently
// running.
func (tp *testProcess) IsRunning() bool {
	process, err := os.FindProcess(tp.Cmd.Process.Pid)
	if err != nil {
		return false
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

// Stop will terminate the associated process. It will attempt a graceful
// shutdown before killing the process.
func (tp *testProcess) Stop() error {
	tp.Cmd.Process.Signal(os.Interrupt) // try shutting down gracefully
	time.Sleep(500 * time.Millisecond)
	if tp.IsRunning() {
		time.Sleep(1 * time.Second)
	} else {
		return nil
	}
	return tp.Cmd.Process.Kill()
}

type gdProcess struct {
	testProcess
	ClientAddress string `toml:"clientaddress"`
	PeerAddress   string `toml:"peeraddress"`
	LocalStateDir string `toml:"localstatedir"`
	RestAuth      bool   `toml:"restauth"`
	Rundir        string `toml:"rundir"`
	uuid          string
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
	args := []string{
		"--config", absConfigFilePath,
		"--localstatedir", g.LocalStateDir,
		"--rundir", g.Rundir,
		"--logdir", path.Join(g.LocalStateDir, "log"),
		"--logfile", "glusterd2.log",
	}
	if externalEtcd {
		args = append(args,
			"--noembed",
			// non-default port to avoid conflict with "real" etcd
			// on system
			// TODO: dynamic ports to avoid test conflicts?
			"--etcdendpoints", "http://localhost:22379",
		)
	}
	g.Cmd = exec.Command(path.Join(binDir, "glusterd2"), args...)

	if err := g.Cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		g.Cmd.Wait()
	}()

	retries := 4
	waitTime := 2000
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

// etcdProcess is used to manage etcd processes by the test suite
// when glusterd2 is not running it's own embedded etcd.
type etcdProcess struct {
	testProcess
	DataDir string
	LogPath string
}

// Spawn starts a new etcd instance.
func (ep *etcdProcess) Spawn() error {
	args := []string{
		"--name", "e2e-test-etcd",
		"--data-dir", ep.DataDir,
		"--listen-client-urls", "http://localhost:22379",
		"--advertise-client-urls", "http://localhost:22379",
		"--log-output", "stdout",
	}
	ep.Cmd = exec.Command("etcd", args...)

	var (
		logf *os.File
		err  error
	)
	if ep.LogPath != "" {
		logf, err = os.OpenFile(ep.LogPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		ep.Cmd.Stdout = logf
		// only the start function needs the open fd. Once the child
		// process has the fd, we want to close the one we opened
		// (on all exit paths).
		defer logf.Close()
	}

	if err := ep.Cmd.Start(); err != nil {
		return err
	}

	go func() {
		ep.Cmd.Wait()
	}()
	// TODO: liveness check?
	return nil
}
