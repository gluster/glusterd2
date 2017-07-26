package main

import (
	"log"
	"os"

	"github.com/gluster/glusterd2/volgen"
	"github.com/gluster/glusterd2/volume"
	"github.com/gluster/glusterd2/xlator"
)

func init() {
	xlator.InitOptions()
}

func main() {
	if len(os.Args) != 3 {
		os.Exit(-1)
	}

	tpath := os.Args[1]
	vpath := os.Args[2]

	gt, err := volgen.ReadTemplateFile(tpath)
	if err != nil {
		log.Fatalf("failed to read and parse template file: %s", err.Error())
	}

	var vol volume.Volinfo
	vol.Name = "test"

	g, err := gt.Generate(&vol)
	if err != nil {
		log.Fatalf("failed to generate graph from template: %s", err.Error())
	}

	err = g.WriteToFile(vpath)
	if err != nil {
		log.Fatalf("failed to write volfile: %s", err.Error())
	}

	log.Printf("generated volfile, %s, from template, %s", vpath, tpath)

	return
}
