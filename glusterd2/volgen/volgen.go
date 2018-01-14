package volgen

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"
)

const (
	fuseTmpl  = "fuse.graph"
	brickTmpl = "brick.graph"
)

var (
	volfilePrefix = "volfiles/"
)

// Generate generates all associated volfiles for the given volinfo.
// NOTE: Currently only does client and brick volfiles
func Generate(vol *volume.Volinfo) error {
	if err := GenerateClientVolfile(vol); err != nil {
		return err
	}

	for _, subvol := range vol.Subvols {
		for _, b := range subvol.Bricks {
			if err := GenerateBrickVolfile(vol, &b); err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateClientVolfile generates the client volfile and stores it in etcd
func GenerateClientVolfile(vol *volume.Volinfo) error {
	ct, err := GetTemplate(fuseTmpl, vol.GraphMap)
	if err != nil {
		return err
	}

	cg, err := ct.Generate(vol, nil)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := cg.Write(buf); err != nil {
		return err
	}
	if _, err := store.Store.Put(context.TODO(), volfilePrefix+vol.Name, buf.String()); err != nil {
		return err
	}

	// XXX: Also write to file, during development
	cg.WriteToFile(getClientVolFilePath(vol.Name))

	return nil
}

// DeleteClientVolfile deletes the client volfile (duh!)
func DeleteClientVolfile(vol *volume.Volinfo) error {

	if _, err := store.Store.Delete(context.TODO(), volfilePrefix+vol.Name); err != nil {
		return err
	}

	// XXX: Also delete the file on disk
	os.Remove(getClientVolFilePath(vol.Name))

	return nil
}

// GenerateBrickVolfile generates the brick volfile for a single brick
func GenerateBrickVolfile(vol *volume.Volinfo, b *brick.Brickinfo) error {
	bt, err := GetTemplate(brickTmpl, vol.GraphMap)
	if err != nil {
		return err
	}

	bg, err := bt.Generate(vol, utils.MergeStringMaps(vol.StringMap(), b.StringMap()))
	if err != nil {
		return err
	}

	return bg.WriteToFile(getBrickVolFilePath(vol.Name, b.NodeID.String(), b.Path))
}

// DeleteBrickVolfile deletes the brick volfile of a single brick
func DeleteBrickVolfile(b *brick.Brickinfo) error {

	path := getBrickVolFilePath(b.VolumeName, b.NodeID.String(), b.Path)
	return os.Remove(path)
}

func getClientVolFilePath(volname string) string {
	dir := utils.GetVolumeDir(volname)
	file := fmt.Sprintf("%s.tcp-fuse.vol", volname)
	return path.Join(dir, file)
}

func getBrickVolFilePath(volname string, brickNodeID string, brickPath string) string {
	dir := utils.GetVolumeDir(volname)
	brickPathWithoutSlashes := strings.Trim(strings.Replace(brickPath, "/", "-", -1), "-")
	file := fmt.Sprintf("%s.%s.%s.vol", volname, brickNodeID, brickPathWithoutSlashes)
	return path.Join(dir, file)
}
