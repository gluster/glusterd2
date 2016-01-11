package etcdmgmt

import (
	"os"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
)

func StartEtcd() error {

	log.Info("Starting etcd")
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not able to get hostname")
		return err
	}

	HostPort2379 := "http://" + hostname + ":2379"

	HostPort2380 := "http://" + hostname + ":2380"

	EtcdStart := exec.Command("/bin/etcd",
		"-listen-client-urls", HostPort2379,
		"-advertise-client-urls", HostPort2379,
		"-listen-peer-urls", HostPort2380,
		"-initial-advertise-peer-urls", HostPort2380,
		"--initial-cluster", "default="+HostPort2380)

	err = EtcdStart.Start()
	if err != nil {
		log.Fatal("Could not start etcd daemon.")
		return err
	}

	// Waiting 1 second for etcd come up. We should not wait for command
	// to finish.
	time.Sleep(1 * time.Second)

	return nil
}
