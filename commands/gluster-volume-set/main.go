package main

import (
	"fmt"

	"github.com/docopt/docopt-go"
	"github.com/kshlm/glusterd2/consul"
	"github.com/kshlm/glusterd2/volume"
)

var (
	usage string = `gluster volume set

Usage:
	gluster-volume-set <volname> (<key> <value>)...
	`
)

func parseVolumeSet() (volname string, keys, values []string) {
	args, _ := docopt.Parse(usage, nil, true, "gluster volume set 0.0", false)

	volname = args["<volname>"].(string)
	keys = args["<key>"].([]string)
	values = args["<value>"].([]string)

	return volname, keys, values
}

func setOptions(v *volume.Volinfo, keys, values []string) bool {
	if len(keys) != len(values) {
		return false
	}

	for i, key := range keys {
		v.Options[key] = values[i]
	}

	return true
}

func main() {
	volname, keys, values := parseVolumeSet()

	c := consul.New()

	if !c.VolumeExists(volname) {
		fmt.Println("Volume", volname, "does not exist.")
		return
	}

	v, err := c.GetVolume(volname)
	if err != nil {
		fmt.Println("An error occurred")
		fmt.Println(err)
		return
	}

	if !setOptions(v, keys, values) {
		fmt.Println("Failed to set options for volume", volname)
		return
	}

	if err := c.AddVolume(v); err != nil {
		fmt.Println("An error occurred")
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully set options for volume", volname)

	return
}
