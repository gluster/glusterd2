package e2e

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

type gdProcess struct {
	Cmd           *exec.Cmd
	ClientAddress string `yaml:"clientaddress"`
	PeerAddress   string `yaml:"peeraddress"`
	Workdir       string `yaml:"workdir"`
	uuid          string
}

func (g *gdProcess) Stop() error {
	return g.Cmd.Process.Kill()
}

func (g *gdProcess) EraseWorkdir() error {
	return os.RemoveAll(g.Workdir)
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

	ubytes, err := ioutil.ReadFile(path.Join(g.Workdir, "uuid"))
	if err != nil {
		return ""
	}
	g.uuid = string(ubytes)
	return g.uuid
}

func (g *gdProcess) IsRestServerUp() bool {
	healthEndpoint := fmt.Sprintf("http://%s/v1/peers/%s/etcdhealth", g.ClientAddress, g.PeerID())
	resp, err := http.Get(healthEndpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}

	return true
}

func spawnGlusterd(configFilePath string) (*gdProcess, error) {

	fContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	g := gdProcess{}
	if err = yaml.Unmarshal(fContent, &g); err != nil {
		return nil, err
	}

	g.EraseWorkdir() // cleanup leftovers from previous test

	if err := os.MkdirAll(path.Join(g.Workdir, "log"), os.ModeDir|os.ModePerm); err != nil {
		return nil, err
	}

	g.Cmd = exec.Command(path.Join(binDir, "glusterd2"),
		"--config", configFilePath,
		"--logdir", path.Join(g.Workdir, "log"),
		"--logfile", "glusterd2.log")

	if err := g.Cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		g.Cmd.Wait()
	}()

	retries := 4
	waitTime := 1000
	for i := 0; i < retries; i++ {
		// opposite of exponential backoff; max sleep time = 1.875s
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
