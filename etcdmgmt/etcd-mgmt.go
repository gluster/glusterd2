package etcdmgmt

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	config "github.com/spf13/viper"
)

// ExecName is to indicate the executable name, useful for mocking in tests
var (
	// Indicates whether the etcd instance is proxy or not
	etcdClient bool
	// arguments used for etcd instance
	listenClientUrls      string
	listenClientProxyUrls string
	advClientUrls         string
	listenPeerUrls        string
	initialAdvPeerUrls    string

	// ETCD executable path
	ExecName = "etcd"
	// Configuration directory for storing etcd configuration
	ETCDConfDir string
	ETCDEnvFile string
	// If this file is touched on ETCDConfDir that indicates that etcd
	// instance need to come with proxy mode
	ETCDProxyFile string

	etcdPidDir, etcdPidFile, etcdLogDir, etcdLogFile string
)

// checkETCDHealth() is to ensure that etcd process have come up properlly or not
func checkETCDHealth(val time.Duration, listenClientUrls string) bool {
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

// StartETCD() is to bring up etcd instance
func StartETCD(args []string) (*os.Process, error) {
	start, pid := isETCDStartNeeded()
	if start == false {
		log.WithField("pid", pid).Info("etcd instance is already running")
		etcdCtx, e := os.FindProcess(pid)
		return etcdCtx, e
	}

	log.WithField("Executable", ExecName).Info("Starting the instance")

	etcdCmd := exec.Command(ExecName, args...)

	utils.InitDir(etcdLogDir)
	outFile, err := os.OpenFile(etcdLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"path": etcdLogFile,
		}).Fatal("Failed to create etcd log")
	}
	etcdCmd.Stdout = outFile
	etcdCmd.Stderr = outFile

	// Don't kill chlid process (etcd) upon ^C (SIGINT) of main glusterd process
	etcdCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err = etcdCmd.Start()
	if err != nil {
		log.WithField("error", err.Error()).Error("Could not start etcd daemon.")
		return nil, err
	}
	// Check the health of node whether etcd have come up properly or not
	var url string
	if etcdClient != true {
		url = args[1]
	} else {
		url = listenClientProxyUrls
	}
	// If its a fresh install and GlusterD is coming up for the first time
	// then check for etcd health, otherwise not
	if gdctx.Restart == false {
		if check := checkETCDHealth(15, url); check != true {
			log.Fatal("Health of etcd is not proper. Check etcd configuration.")
		}
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
		log.WithField("err", err).Error("Failed to write etcd pid")
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
//   is found then check whether etcd health is ok, if so then return true else
//   false
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
			listenClientUrls := "http://" + gdctx.HostIP + ":2379"
			_, err = http.Get(listenClientUrls + "/health")
			if err != nil {
				log.WithField("err", err).Error("etcd health check failed")
				pid = -1
				start = true
			} else {
				start = false
			}
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

// initETCDArgVar() will initialize etcd arguments and variables which will be used at various places
func initETCDArgVar() {
	// Set the paths required for etcd
	ETCDConfDir = path.Join(config.GetString("localstatedir"), "etcd")
	ETCDEnvFile = path.Join(ETCDConfDir, "etcdenv.conf")
	ETCDProxyFile = path.Join(ETCDConfDir, "proxy")

	etcdPidDir = path.Join(config.GetString("rundir"), "etcd")
	etcdPidFile = path.Join(etcdPidDir, "etcd.pid")
	etcdLogDir = path.Join(config.GetString("logdir"), "etcd")
	etcdLogFile = path.Join(etcdLogDir, "etcd.log")

	gdctx.SetLocalHostIP()

	listenClientUrls = "http://" + gdctx.HostIP + ":2379"
	listenClientProxyUrls = listenClientUrls
	advClientUrls = "http://" + gdctx.HostIP + ":2379"
	listenPeerUrls = "http://" + gdctx.HostIP + ":2380"
	initialAdvPeerUrls = "http://" + gdctx.HostIP + ":2380"
}

// formETCDArgs constructs the arguments to be passed to etcd binary
func formETCDArgs() []string {
	etcdClient = false
	var args []string
	m := make(map[string]string)

	// If proxy file doesn't exist then etcd to come up as server else
	// client
	f, err := os.Open(ETCDProxyFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Debug("etcd instance will come up with server mode")
		default:
			log.WithFields(log.Fields{
				"error": err,
				"path":  ETCDProxyFile,
			}).Fatal("Failed to read from file")
		}
	} else {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			linestr := scanner.Text()
			r := strings.Split(linestr, "=")

			k := r[0]
			v := r[1]
			l := len(r)
			for i := 2; i < l; i++ {
				v = v + "=" + r[i]
			}
			m[k] = v
		}

		log.Debug("etcd instance will come up with proxy mode")
		etcdClient = true
	}
	if etcdClient == true {
		args = []string{"-proxy", "on",
			"-listen-client-urls", listenClientProxyUrls,
			"-initial-cluster", m["ETCD_INITIAL_CLUSTER"]}
	} else {
		args = []string{"-listen-client-urls", listenClientUrls,
			"-advertise-client-urls", advClientUrls,
			"-listen-peer-urls", listenPeerUrls,
			"-initial-advertise-peer-urls", initialAdvPeerUrls}
	}
	return args
}

// ETCDStartInit() Check whether etcd environment variable present or not
// If it present then start etcd without --initial-cluster flag
// other wise start etcd normally.
func ETCDStartInit() (*os.Process, error) {
	initETCDArgVar()

	f, err := os.Open(ETCDEnvFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Info("Starting/Restarting etcd for a initial node")
			_, e := os.Stat(ETCDProxyFile)
			if e == nil {
				etcdClient = true
			}
			return StartStandAloneETCD()
		default:
			log.WithFields(log.Fields{
				"error": err,
				"path":  ETCDEnvFile,
			}).Fatal("Failed to read from file")
		}
	} else {
		defer f.Close()

		// Restoring etcd environment variable and starting etcd daemon
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			linestr := scanner.Text()
			etcdenv := strings.Split(linestr, "=")

			etcdEnvKey := etcdenv[0]
			etcdEnvData := etcdenv[1]
			envlen := len(etcdenv)
			for i := 2; i < envlen; i++ {
				etcdEnvData = etcdEnvData + "=" + etcdenv[i]
			}

			// setting etcd environment variable
			os.Setenv(etcdEnvKey, etcdEnvData)
		}
		args := formETCDArgs()
		log.Info("Starting etcd daemon")
		return StartETCD(args)
	}
	return nil, err
}

//StartStandAloneETCD() will Start default etcd by considering single server node
func StartStandAloneETCD() (*os.Process, error) {
	var args []string
	if etcdClient == true {
		args = formETCDArgs()
	} else {
		args = []string{"-listen-client-urls", listenClientUrls,
			"-advertise-client-urls", advClientUrls,
			"-listen-peer-urls", listenPeerUrls,
			"-initial-advertise-peer-urls", initialAdvPeerUrls,
			"-initial-cluster", "default=" + listenPeerUrls}
	}

	return StartETCD(args)
}

// StopETCD() will Stop etcd process on the node
func StopETCD(etcdCtx *os.Process) error {
	err := etcdCtx.Kill()
	if err != nil {
		log.Error("Could not able to kill etcd daemon")
		return err
	}
	_, err = etcdCtx.Wait()
	if err != nil {
		log.Error("Could not able to kill etcd daemon")
		return err
	}
	log.Debug("Stopped a running etcd instance")
	return nil
}

// ReStartETCD() will restart etcd process
func ReStartETCD() (*os.Process, error) {
	// Stop etcd process
	etcdCtx := gdctx.EtcdProcessCtx
	err := StopETCD(etcdCtx)
	if err != nil {
		log.Error("Could not able to stop etcd daemon")
		return nil, err
	}
	args := formETCDArgs()
	log.WithField("args", args).Info("Restarting etcd daemon")
	if etcdClient == true {
		log.Info("Removing old etcd configuration")
		//TODO : this is hardcoded now, with -data-dir coming in this
		//needs to be configurable
		os.RemoveAll("/default.etcd")
	}
	return StartETCD(args)
}
