package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	glusterd1Workdir = "/var/lib/glusterd"
	volumeStatus     = map[string]string{
		"0": "Created",
		"1": "Started",
		"2": "Stopped",
	}
	volumeTypes = map[string]string{
		"0": "Distribute",
		"1": "Replicate",
		"2": "Distributed-Replicate",
		"3": "Disperse",
		"4": "Distributed-Disperse",
	}
)

func handleErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

type Brickinfo struct {
	PeerID   string `json:"peer-id"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
}

type Subvolinfo struct {
	Bricks []Brickinfo `json:"bricks"`
	Type   string      `json:"type"`
}

type Volinfo struct {
	Name    string       `json:"name"`
	ID      string       `json:"id"`
	Type    string       `json:"type"`
	Status  string       `json:"status"`
	Subvols []Subvolinfo `json:"subvols"`
}

func main() {
	files, err := ioutil.ReadDir(path.Join(glusterd1Workdir, "vols"))
	handleErr(err)

	var allVolumesData []Volinfo
	for _, f := range files {
		v := f.Name()
		volumeData := make(map[string]string)
		info, err := ioutil.ReadFile(path.Join(glusterd1Workdir, "vols", v, "info"))
		handleErr(err)
		data := strings.Trim(string(info), "\n")
		for _, d := range strings.Split(data, "\n") {
			parts := strings.SplitN(d, "=", 2)
			volumeData[parts[0]] = parts[1]
		}

		totalBricks, err := strconv.Atoi(volumeData["count"])
		replicaCount, err := strconv.Atoi(volumeData["replica_count"])
		disperseCount, err := strconv.Atoi(volumeData["disperse_count"])

		subvolSize := totalBricks
		subvolType := "Distribute"
		if replicaCount > 1 {
			subvolSize = replicaCount
			subvolType = "Replicate"
		} else if disperseCount > 1 {
			subvolSize = disperseCount
			subvolType = "Disperse"
		}

		numSubvols := totalBricks / subvolSize

		volinfo := Volinfo{
			Name:    v,
			ID:      volumeData["volume-id"],
			Status:  volumeStatus[volumeData["status"]],
			Type:    volumeTypes[volumeData["type"]],
			Subvols: []Subvolinfo{},
		}

		// Bricks parse
		for sv := 0; sv < numSubvols; sv++ {
			subvol := Subvolinfo{
				Type:   subvolType,
				Bricks: []Brickinfo{},
			}

			for bi := 0; bi < subvolSize; bi++ {
				idx := fmt.Sprintf("brick-%d", sv*subvolSize+bi)
				brickparts := strings.Split(
					strings.Replace(volumeData[idx], "-", "/", -1),
					":",
				)
				subvol.Bricks = append(subvol.Bricks,
					Brickinfo{
						Hostname: brickparts[0],
						Path:     brickparts[1],
					},
				)
			}
			volinfo.Subvols = append(volinfo.Subvols, subvol)
		}
		allVolumesData = append(allVolumesData, volinfo)
	}

	out, err := json.Marshal(allVolumesData)
	handleErr(err)

	fmt.Println(string(out))
}
