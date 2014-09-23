package main

import (
	"fmt"
	"strconv"

	"github.com/docopt/docopt-go"
	"github.com/kshlm/glusterd2/consul"
	"github.com/kshlm/glusterd2/volume"
)

var (
	usage string = `Create Volume

Usage:
	create-volume <volname> [options] <brick>...
Options:
	--transport=<transport>          Transport type (tcp|rdma|tcp,rdma) [default: tcp]
	--replica=<replica-count>        Replication count [default: 1]
	--stripe=<stripe-count>          Stripe count [default: 1]
	--disperse=<disperse-count>      Disperse count [default: 1]
	--redundancy=<redundancy-count>  Redundancy count [default: 1]
	--force

	`
)

func parseCreateVolume() *volume.Volinfo {
	args, _ := docopt.Parse(usage, nil, true, "Gluster-create-volume 1.0", false)

	volname := args["<volname>"].(string)
	transport := args["--transport"].(string)
	replica, _ := strconv.Atoi(args["--replica"].(string))
	stripe, _ := strconv.Atoi(args["--stripe"].(string))
	disperse, _ := strconv.Atoi(args["--disperse"].(string))
	redundancy, _ := strconv.Atoi(args["--redundancy"].(string))
	bricks := args["<brick>"].([]string)

	v := volume.New(volname, transport, uint16(replica), uint16(stripe), uint16(disperse), uint16(redundancy), bricks)

	return v
}

func main() {
	v := parseCreateVolume()
	c := consul.New()

	if c.VolumeExists(v.Name) {
		fmt.Println("Volume with name", v.Name, "exists")
		return
	}

	err := c.AddVolume(v)
	if err != nil {
		fmt.Println("Failed to create volume", v.Name)
	} else {
		fmt.Println("Successfully created volume", v.Name)
	}
}
