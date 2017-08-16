package volgen

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	config "github.com/spf13/viper"
)

// TODO: differentiate between various types of client volfiles
var volfilePrefix = store.GlusterPrefix + "volfiles/"

// TODO: This is a quick and dirty reference implementation that should
// be replaced when real volgen with dependency resolution is ready.
// This is not complete either - works only for dist, rep and dist-rep
// volumes.

// GenerateClientVolfile generates the client volfile and stores it in etcd
func GenerateClientVolfile(vinfo *volume.Volinfo) error {

	volfile := new(bytes.Buffer)

	// Insert leaf nodes i.e client xlators
	for index, b := range vinfo.Bricks {

		address, err := utils.FormRemotePeerAddress(b.Hostname)
		if err != nil {
			return err
		}
		remoteHost, _, _ := net.SplitHostPort(address)

		replacer := strings.NewReplacer(
			"<child-index>", strconv.Itoa(index),
			"<brick-path>", b.Path,
			"<volume-name>", vinfo.Name,
			"<trusted-username>", vinfo.Auth.Username,
			"<trusted-password>", vinfo.Auth.Password,
			"<remote-host>", remoteHost)

		volfile.WriteString(replacer.Replace(clientLeafTemplate))
	}

	var subvols []string

	wbSubvol := "replicate"
	// Create AFR xlator entries
	if vinfo.ReplicaCount > 1 {
		bindex := 0
		afrInstanceCount := len(vinfo.Bricks) / vinfo.ReplicaCount
		for rindex := 0; rindex < afrInstanceCount; rindex++ {
			subvols = make([]string, vinfo.ReplicaCount)
			for j := 0; j < vinfo.ReplicaCount; j++ {
				subvols[j] = fmt.Sprintf("%s-client-%s", vinfo.Name, strconv.Itoa(bindex))
				bindex++
			}
			var childIndex string
			if afrInstanceCount == 1 {
				childIndex = ""
			} else {
				childIndex = "-" + strconv.Itoa(rindex)
			}
			replacer := strings.NewReplacer(
				"<volume-name>", vinfo.Name,
				"<afr-pending-xattr>", strings.Join(subvols, ","),
				"<afr-subvolumes>", strings.Join(subvols, " "),
				"<child-index>", childIndex)
			volfile.WriteString(replacer.Replace(clientVolfileAFRTemplate))
		}
	}

	// Create DHT xlator entry
	if (vinfo.ReplicaCount != len(vinfo.Bricks)) || (len(vinfo.Bricks) == 1) {
		wbSubvol = "dht"
		if vinfo.ReplicaCount > 1 {
			// AFR instances are children of DHT (dist-rep)
			afrInstanceCount := len(vinfo.Bricks) / vinfo.ReplicaCount
			subvols = make([]string, afrInstanceCount)
			for aindex := 0; aindex < afrInstanceCount; aindex++ {
				subvols[aindex] = fmt.Sprintf("%s-replicate-%s", vinfo.Name, strconv.Itoa(aindex))
			}
		} else {
			// Client xlators are children of DHT (pure distribute)
			subvols = make([]string, len(vinfo.Bricks))
			for bindex := 0; bindex < len(vinfo.Bricks); bindex++ {
				subvols[bindex] = fmt.Sprintf("%s-client-%s", vinfo.Name, strconv.Itoa(bindex))
			}
		}
		replacer := strings.NewReplacer(
			"<volume-name>", vinfo.Name,
			"<dht-subvolumes>", strings.Join(subvols, " "))
		volfile.WriteString(replacer.Replace(clientVolfileDHTTemplate))
	}

	// Insert all other xlator entries (linear graph here onwards)
	replacer := strings.NewReplacer("<volume-name>", vinfo.Name, "<wb-subvol>", wbSubvol)
	volfile.WriteString(replacer.Replace(clientVolfileBaseTemplate))

	if _, err := store.Store.Put(context.TODO(), volfilePrefix+vinfo.Name, volfile.String()); err != nil {
		return err
	}

	return nil
}

// DeleteClientVolfile deletes the client volfile (duh!)
func DeleteClientVolfile(vol *volume.Volinfo) error {

	if _, err := store.Store.Delete(context.TODO(), volfilePrefix+vol.Name); err != nil {
		return err
	}

	return nil
}

func getBrickVolFilePath(volumeName string, brickNodeID string, brickPath string) string {
	volumeDir := utils.GetVolumeDir(volumeName)
	brickPathWithoutSlashes := strings.Trim(strings.Replace(brickPath, "/", "-", -1), "-")
	volFileName := fmt.Sprintf("%s.%s.%s.vol", volumeName, brickNodeID, brickPathWithoutSlashes)
	return path.Join(volumeDir, volFileName)
}

// GenerateBrickVolfile generates the brick volfile for a single brick
func GenerateBrickVolfile(vinfo *volume.Volinfo, binfo *brick.Brickinfo) error {

	bpath := getBrickVolFilePath(vinfo.Name, binfo.NodeID.String(), binfo.Path)
	f, err := os.Create(bpath)
	if err != nil {
		return err
	}
	defer f.Close()

	replacer := strings.NewReplacer(
		"<volume-name>", vinfo.Name,
		"<volume-id>", vinfo.ID.String(),
		"<brick-path>", binfo.Path,
		"<trusted-username>", vinfo.Auth.Username,
		"<trusted-password>", vinfo.Auth.Password,
		"<local-state-dir>", config.GetString("localstatedir"))

	if _, err = replacer.WriteString(f, brickVolfileTemplate); err != nil {
		return err
	}
	f.Sync()

	return nil
}

// DeleteBrickVolfile deletes the brick volfile of a single brick
func DeleteBrickVolfile(binfo *brick.Brickinfo) error {

	path := getBrickVolFilePath(binfo.VolumeName, binfo.NodeID.String(), binfo.Path)
	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}
