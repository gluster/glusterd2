//Core file for graph and volfile generation

package volgen

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gluster/glusterd2/brick"
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

// GenerateVolfile function will do all task from graph generation to volfile generation
func GenerateVolfile(vinfo *volume.Volinfo, vauth *volume.VolAuth) error {

	// Create 'vols' directory.
	err := os.MkdirAll(utils.GetVolumeDir(vinfo.Name), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	// Generate client volfile
	graph := GenerateGraph(vinfo, "CLIENT")
	cpath := getClientVolFilePath(vinfo)
	f, err := os.Create(cpath)
	if err != nil {
		return err
	}
	graph.DumpGraph(f)
	f.Close()

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

		replacer := strings.NewReplacer(
			"<volume-name>", vinfo.Name,
			"<volume-id>", vinfo.ID.String(),
			"<brick-path>", b.Path,
			"<trusted-username>", vauth.Username,
			"<trusted-password>", vauth.Password)

		_, err = replacer.WriteString(f, brick.VolfileTemplate)
		if err != nil {
			return err
		}

		f.Close()
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
