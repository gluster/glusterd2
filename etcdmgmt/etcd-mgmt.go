package etcdmgmt

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

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

	//Checking health of etcd cluster
	for {
		result := struct{ Health string }{}

		resp, err := http.Get(HostPort2379 + "/health")
		if err != nil {
			continue
		}

		Body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal([]byte(Body), &result)
		if err != nil {
			continue
		}
		if result.Health == "true" {
			break
		}
	}

	return nil
}
