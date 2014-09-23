package main

import (
	"fmt"

	"github.com/docopt/docopt-go"
	"github.com/kshlm/glusterd2/consul"
)

var (
	usage string = `gluster volume info

Usage:
	get-volume <volname>
	`
)

func parseVolumeInfo() string {
	args, _ := docopt.Parse(usage, nil, true, "gluster volume info 0.0", false)

	volname := args["<volname>"].(string)
	return volname
}

func main() {
	name := parseVolumeInfo()
	c := consul.New()

	vol, err := c.GetVolume(name)
	if err != nil || vol == nil {
		fmt.Println("Volume", name, "does not exist")
		return
	}

	fmt.Println(vol)
}
