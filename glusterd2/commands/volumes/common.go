package volumecommands

import (
	"os"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

func areOptionNamesValid(optsFromReq map[string]string) error {

	for o := range optsFromReq {
		_, err := options.Find(o)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateBrickVolfiles(c transaction.TxnCtx) error {

	// This is used in volume-create and volume-set

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	// Create 'vols' directory.
	err := os.MkdirAll(utils.GetVolumeDir(volinfo.Name), os.ModeDir|os.ModePerm)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateBrickVolfiles: failed to create vol directory")
		return err
	}

	for _, b := range volinfo.Bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}
		if err := volgen.GenerateBrickVolfile(&volinfo, &b); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("generateBrickVolfiles: failed to create brick volfile")
			return err
		}
	}

	return nil
}

func notifyVolfileChange(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if volinfo.State != volume.VolStarted {
		return nil
	}

	sunrpc.FetchSpecNotify(c)

	return nil
}

func storeVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volgen.GenerateClientVolfile(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateVolfiles: failed to create client volfile")
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}

	return nil
}
