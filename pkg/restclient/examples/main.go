package main

import (
	"fmt"
	"github.com/gluster/glusterd2/restclient"
)

const (
	baseURL  = "http://localhost:24007"
	username = ""
	password = ""
	peerNode = "node2"
	volname  = "gv1"
	brick1   = "10.70.1.111:/bricks/b1"
	brick2   = "10.70.1.111:/bricks/b2"
	force    = true
	replica  = 0
)

func main() {
	restClient := restclient.NewRESTClient(baseURL, username, password)
	fmt.Println(restClient.PeerProbe(peerNode))
	fmt.Println(restClient.PeerDetach(peerNode))
	fmt.Println(restClient.VolumeCreate(volname, []string{brick1, brick2}, replica, force))
	fmt.Println(restClient.VolumeStart(volname))
	fmt.Println(restClient.VolumeStop(volname))
	fmt.Println(restClient.VolumeDelete(volname))
}
