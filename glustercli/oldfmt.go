package main

import (
	"os"
)

var newOptions = map[string]string{
	"force": "--force",
}

var newOptionsVolume = map[string]string{
	"replica":       "--replica",
	"arbiter":       "--arbiter",
	"disperse":      "--disperse",
	"disperse-data": "--disperse-data",
	"redundancy":    "--redundancy",
	"transport":     "--transport",
}

func argsMigrate() {
	for k, arg := range os.Args {
		if val, ok := newOptions[arg]; ok {
			os.Args[k] = val
		}
	}

	if len(os.Args) >= 3 && os.Args[1] == "volume" && os.Args[2] == "create" {
		for k, arg := range os.Args {
			if val, ok := newOptionsVolume[arg]; ok {
				os.Args[k] = val
			}
		}
	}
}
