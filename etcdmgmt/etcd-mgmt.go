package etcdmgmt

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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

	listenClientUrls := "http://" + hostname + ":2379"

	advClientUrls := "http://" + hostname + ":2379"

	listenPeerUrls := "http://" + hostname + ":2380"

	initialAdvPeerUrls := "http://" + hostname + ":2380"

	etcdStart := exec.Command("/bin/etcd",
		"-listen-client-urls", listenClientUrls,
		"-advertise-client-urls", advClientUrls,
		"-listen-peer-urls", listenPeerUrls,
		"-initial-advertise-peer-urls", initialAdvPeerUrls,
		"--initial-cluster", "default="+listenPeerUrls)

	err = etcdStart.Start()
	if err != nil {
		log.Fatal("Could not start etcd daemon.", err)
		return err
	}

	result := struct{ Health string }{}
	// Checking health of etcd. Health of the etcd should be true,
	// means etcd have initialized properly before using any etcd command
	for {

		// Waiting for 15 second. Within 15 second health of etcd should
		// be true otherwise it should throw an error
		timer := time.NewTimer(time.Second * 15)
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

	return nil
}
