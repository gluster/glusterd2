package etcdmgmt

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
)

// ExecName is to indicate the executable name, useful for mocking in tests
var ExecName = "etcd"

func checkHealth(val time.Duration, listenClientUrls string) bool {
	result := struct{ Health string }{}
	// Checking health of etcd. Health of the etcd should be true,
	// means etcd have initialized properly before using any etcd command
	timer := time.NewTimer(time.Second * val)
	for {
		// Waiting for 15 second. Within 15 second health of etcd should
		// be true otherwise it should throw an error
		go func() {
			<-timer.C
			if result.Health != "true" {
				log.Fatal("Health of etcd is not proper. Check etcd configuration.")
			}
		}()

		resp, err := http.Get(listenClientUrls + "/health")
		if err != nil {
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal([]byte(body), &result)
		if err != nil {
			continue
		}
		if result.Health == "true" {
			timer.Stop()
			break
		}
	}
	return true
}

var (
	etcdPidDir  = "/var/run/gluster/"
	etcdPidFile = etcdPidDir + "etcd.pid"
)

// StartETCD () is to bring up etcd instance
func StartETCD() (*os.Process, error) {
	start, pid := isETCDStartNeeded()
	if start == false {
		log.WithField("pid", pid).Info("etcd instance is already running")
		etcdCtx, e := os.FindProcess(pid)
		return etcdCtx, e
	}

	log.WithField("Executable", ExecName).Info("Starting")
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not able to get hostname")
		return nil, err
	}

	listenClientUrls := "http://" + hostname + ":2379"

	advClientUrls := "http://" + hostname + ":2379"

	listenPeerUrls := "http://" + hostname + ":2380"

	initialAdvPeerUrls := "http://" + hostname + ":2380"

	etcdCmd := exec.Command(ExecName,
		"-listen-client-urls", listenClientUrls,
		"-advertise-client-urls", advClientUrls,
		"-listen-peer-urls", listenPeerUrls,
		"-initial-advertise-peer-urls", initialAdvPeerUrls,
		"--initial-cluster", "default="+listenPeerUrls)

	// TODO: use unix.Setpgid instead of using syscall
	// Don't kill chlid process (etcd) upon ^C (SIGINT) of main glusterd process
	etcdCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err = etcdCmd.Start()
	if err != nil {
		log.WithField("error", err.Error()).Error("Could not start etcd daemon.")
		return nil, err
	}

	if check := checkHealth(15, listenClientUrls); check != true {
		log.Fatal("Health of etcd is not proper. Check etcd configuration.")
	}
	log.WithField("pid", etcdCmd.Process.Pid).Debug("etcd pid")
	if err := writeETCDPidFile(etcdCmd.Process.Pid); err != nil {
		etcdCmd.Process.Kill()
		return nil, err
	}
	return etcdCmd.Process, nil
}

// writeETCDPidFile () is to write the pid of etcd instance
func writeETCDPidFile(pid int) error {
	// create directory to store etcd pid if it doesn't exist
	utils.InitDir(etcdPidDir)
	if err := ioutil.WriteFile(etcdPidFile, []byte(strconv.Itoa(pid)), os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdPidFile,
			"pid":   string(pid),
		}).Error("Failed to write etcd pid to the file")
		return err
	}
	return nil
}

// isETCDStartNeeded() reads etcd.pid file
// @ if pid is not found returns true
// @ if pid is found, checks for the process with the pid, if a running instance
//   is found return false else true

func isETCDStartNeeded() (bool, int) {
	pid := -1
	start := true
	bytes, err := ioutil.ReadFile(etcdPidFile)
	if err == nil {
		pidString := string(bytes)
		if pid, err = strconv.Atoi(pidString); err != nil {
			log.WithField("pid", pidString).Error("Failed to convert string to integer")
			start = true
			return start, pid
		}

		if exist := utils.CheckProcessExist(pid); exist == true {
			start = false
		}
	} else {
		switch {
		case os.IsNotExist(err):
			start = true
			break
		default:
			log.WithFields(log.Fields{
				"error": err,
				"path":  etcdPidFile,
			}).Fatal("Failed to read from file")
		}
	}
	return start, pid
}
