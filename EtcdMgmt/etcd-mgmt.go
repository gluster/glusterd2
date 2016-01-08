package EtcdMgmt

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

	Ip2379 := "http://" + hostname + ":2379"

	Ip2380 := "http://" + hostname + ":2380"

	etcd_start := exec.Command("/bin/etcd",
		"-listen-client-urls", Ip2379,
		"-advertise-client-urls", Ip2379,
		"-listen-peer-urls", Ip2380,
		"-initial-advertise-peer-urls", Ip2380,
		"--initial-cluster", "default="+Ip2380)

	err = etcd_start.Start()
	if err != nil {
		log.Fatal("Could not start etcd daemon.")
		return err
	}
	// Waiting 1 second for etcd come up. We should not wait for command
	// to finish.
	time.Sleep(1 * time.Second)

	return nil
}
