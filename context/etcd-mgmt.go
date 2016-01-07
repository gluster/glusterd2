package context

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func StartEtcd() {

	log.Info("Starting etcd")
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not able to get hostname")
	}

	ip_2379 := strings.Join([]string{"http://", ":2379"}, hostname)

	ip_2380 := strings.Join([]string{"http://", ":2380"}, hostname)

	default_etcd := strings.Join([]string{"default=", ""}, ip_2380)

	etcd_start := exec.Command("/bin/etcd",
		"-listen-client-urls", ip_2379,
		"-advertise-client-urls", ip_2379,
		"-listen-peer-urls", ip_2380,
		"-initial-advertise-peer-urls", ip_2380,
		"--initial-cluster", default_etcd)

	err = etcd_start.Start()
	if err != nil {
		log.Fatal("Could not start etcd daemon.")
	}
}
