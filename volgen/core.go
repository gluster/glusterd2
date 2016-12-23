package volgen

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

var (
	// GenerateVolfileFunc will do all task from graph generation to volfile generation
	GenerateVolfileFunc = GenerateVolfile
)

func buildClientVolfile(vinfo *volume.Volinfo, vauth *volume.VolAuth) (*bytes.Buffer, error) {
	// TODO: This is a quick and dirty reference implementation that should
	// be replaced when real volgen with dependency resolution is ready.
	// This is not complete either - works only for dist, rep and dist-rep
	// volumes.

	if len(vinfo.Bricks)%vinfo.ReplicaCount != 0 {
		// For replicate, distributed or distributed-replicated:
		// No. of bricks must be a multiple of replica count.
		return nil, errors.New("Brick count should be a multiple of replica count")
	}

	volfile := new(bytes.Buffer)

	// Insert leaf nodes i.e client xlators
	for index, b := range vinfo.Bricks {

		address, err := utils.FormRemotePeerAddress(b.Hostname)
		if err != nil {
			return nil, err
		}
		remoteHost, _, _ := net.SplitHostPort(address)

		replacer := strings.NewReplacer(
			"<child-index>", strconv.Itoa(index),
			"<brick-path>", b.Path,
			"<volume-name>", vinfo.Name,
			"<trusted-username>", vauth.Username,
			"<trusted-password>", vauth.Password,
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
	if vinfo.ReplicaCount != len(vinfo.Bricks) {
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

	return volfile, nil
}

// GenerateVolfile function will do all task from graph generation to volfile generation
func GenerateVolfile(vinfo *volume.Volinfo, vauth *volume.VolAuth) error {

	// Create 'vols' directory.
	err := os.MkdirAll(utils.GetVolumeDir(vinfo.Name), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// Generate client volfile
	cpath := getClientVolFilePath(vinfo)
	volfileContent, err := buildClientVolfile(vinfo, vauth)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cpath, volfileContent.Bytes(), 0644)
	if err != nil {
		return err
	}

	// Generate brick volfiles
	for _, b := range vinfo.Bricks {

		// Generate brick volfiles for only those bricks that belong
		// to this node/instance.
		if !uuid.Equal(b.ID, gdctx.MyUUID) {
			continue
		}

		bpath := utils.GetBrickVolFilePath(vinfo.Name, b.Hostname, b.Path)

		f, err := os.Create(bpath)
		if err != nil {
			return err
		}
		defer f.Close()

		replacer := strings.NewReplacer(
			"<volume-name>", vinfo.Name,
			"<volume-id>", vinfo.ID.String(),
			"<brick-path>", b.Path,
			"<trusted-username>", vauth.Username,
			"<trusted-password>", vauth.Password)

		_, err = replacer.WriteString(f, brickVolfileTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func getClientVolFilePath(vinfo *volume.Volinfo) string {
	volfileName := fmt.Sprintf("trusted-%s.tcp-fuse.vol", vinfo.Name)
	return path.Join(utils.GetVolumeDir(vinfo.Name), volfileName)
}

// DeleteVolfile deletes the volfiles created for the volume
func DeleteVolfile(vol *volume.Volinfo) error {

	// TODO: It would be much simpler to remove whole volume directory here ?

	path := getClientVolFilePath(vol)
	err := os.Remove(path)
	if err != nil {
		// TODO: log using txn logger context instead
		log.WithFields(log.Fields{
			"path":  path,
			"error": err,
		}).Error("DeleteVolfile: Failed to remove client volfile")
		return err
	}

	for _, b := range vol.Bricks {

		if !uuid.Equal(b.ID, gdctx.MyUUID) {
			continue
		}

		path := utils.GetBrickVolFilePath(vol.Name, b.Hostname, b.Path)
		err := os.Remove(path)
		if err != nil {
			// TODO: log using txn logger context instead
			log.WithFields(log.Fields{
				"path":  path,
				"error": err,
			}).Error("DeleteVolfile: Failed to remove brick volfile")
			return err
		}
	}

	return nil
}
