//Core file for graph and volfile generation

package volgen

import (
	"fmt"
	"os"
	"strings"

	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"
)

var (
	// GenerateVolfileFunc will do all task from graph generation to volfile generation
	GenerateVolfileFunc = GenerateVolfile
)

// GenerateVolfile function will do all task from graph generation to volfile generation
func GenerateVolfile(vinfo *volume.Volinfo) error {
	var cpath string
	var err error
	var f *os.File

	graph := GenerateGraph(vinfo, "CLIENT")

	//Generate client volfile
	err = getClientFilePath(vinfo, &cpath)
	if err != nil {
		return err
	}

	f, err = os.Create(cpath)
	if err != nil {
		return err
	}
	graph.DumpGraph(f)
	f.Close()

	// Generate brick volfiles
	for _, b := range vinfo.Bricks {

		// Create 'vols' directory.
		vdir := utils.GetVolumeDir(vinfo.Name)
		err := os.MkdirAll(vdir, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}

		// Get brick volfile path
		cpath = utils.GetBrickVolFilePath(vinfo.Name, b.Hostname, b.Path)

		f, err = os.Create(cpath)
		if err != nil {
			return err
		}

		replacer := strings.NewReplacer(
			"<volume-name>", vinfo.Name,
			"<volume-id>", vinfo.ID.String(),
			"<brick-path>", b.Path)

		_, err = replacer.WriteString(f, brick.VolfileTemplate)
		if err != nil {
			return err
		}

		f.Close()
	}
	return nil
}

// Fix this: Move it to utils and make it return string
func getClientFilePath(vinfo *volume.Volinfo, path *string) error {
	vdir := utils.GetVolumeDir(vinfo.Name)

	// Create volume directory (/var/lib/glusterd/vols/<VOLNAME>)
	err := os.MkdirAll(vdir, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	*path = fmt.Sprintf("%s/trusted-%s.tcp-fuse.vol", vdir, vinfo.Name)
	return err
}

// DeleteVolfile deletes the volfiles created for the volume
// XXX: This is a quick and dirty implementation with no error checking.
// A proper implementation will be implemented when volgen is re-implemented.
func DeleteVolfile(vol *volume.Volinfo) error {
	var path string

	_ = getClientFilePath(vol, &path)
	_ = os.Remove(path)

	for _, b := range vol.Bricks {
		path := utils.GetBrickVolFilePath(vol.Name, b.Hostname, b.Path)
		_ = os.Remove(path)
	}
	return nil
}
