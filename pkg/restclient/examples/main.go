package main

import (
	"fmt"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/restclient"
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
	client := restclient.New(baseURL, username, password)
	fmt.Println(client.PeerProbe(peerNode))
	fmt.Println(client.PeerDetach(peerNode))
	req := api.VolCreateReq{
		Name:    volname,
		Bricks:  []string{brick1, brick2},
		Replica: replica,
		Force:   force,
	}
	fmt.Println(client.VolumeCreate(req))
	fmt.Println(client.VolumeStart(volname))
	fmt.Println(client.VolumeStop(volname))
	fmt.Println(client.VolumeDelete(volname))
}
