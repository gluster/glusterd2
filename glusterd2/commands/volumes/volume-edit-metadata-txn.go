package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
)

func editVolMetadata(c transaction.TxnCtx) error {

	var req api.VolEditMetadataReq
	var volname string

	if err := c.Get("req", &req); err != nil {
		return err
	}

	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Debug("Failed to get key from transaction context.")
		return err
	}

	v, err := volume.GetVolume(volname)
	if err != nil {
		c.Logger().WithError(err).Error("Failed to get volinfo from the store")
		return err
	}

	for key, _ := range req.Metadata {
		v.Metadata[key] = req.Metadata[key]
	}

	if err := c.Set("volinfo", v); err != nil {
		return err
	}

	return nil
}
