package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"

	toml "github.com/pelletier/go-toml"
)

type gdProcess struct {
	Cmd           *exec.Cmd
	ClientAddress string `toml:"clientaddress"`
	PeerAddress   string `toml:"peeraddress"`
	Workdir       string `toml:"workdir"`
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

	// Endpoint doesn't matter here. All responses include a
	// X-Gluster-Node-Id response header.
	endpoint := fmt.Sprintf("http://%s/version", g.ClientAddress)
	resp, err := http.Get(endpoint)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	g.uuid = resp.Header.Get("X-Gluster-Node-Id")
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

	if cleanStart {
		g.EraseWorkdir() // cleanup leftovers from previous test
	}

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
			p.EraseWorkdir()
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
		gd.EraseWorkdir()
	}
	return nil
}

func initRestclient(clientAddress string) *restclient.Client {
	return restclient.New("http://"+clientAddress, "", "", "", false)
}
