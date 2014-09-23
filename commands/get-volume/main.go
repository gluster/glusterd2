package main

import (
	"fmt"

	"github.com/docopt/docopt-go"
	"github.com/kshlm/glusterd2/consul"
)

var (
	usage string = `Create Volume

Usage:
	get-volume <volname>
	`
)

func parseGetVolume() string {
	args, _ := docopt.Parse(usage, nil, true, "Gluster-create-volume 1.0", false)

	volname := args["<volname>"].(string)
	return volname
}

func main() {
	name := parseGetVolume()
	c := consul.New()

	vol, err := c.GetVolume(name)
	if err != nil {
		fmt.Println("Volume", name, "does not exist")
		fmt.Println(err)
		return
	}

	fmt.Println(vol)
}
