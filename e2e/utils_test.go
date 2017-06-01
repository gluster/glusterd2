package e2e

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"
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
		err := g.Cmd.Wait()
		log.WithFields(log.Fields{
			"pid":    g.Cmd.Process.Pid,
			"status": err,
		}).Error("Child exited.")
	}()

	return &g, nil
}
