package devicecommands

import (
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var deviceinfo api.Info
	
	if err := c.Get("peerid", &deviceinfo.PeerID); err != nil {
		log.WithField("error", err).Error("Failed transaction, cannot find peerid")
		return err
	}
	if err := c.Get("names", &deviceinfo.Names); err != nil {
		log.WithField("error", err).Error("Failed transaction, cannot find device names")
		return err
	}
	for _, element := range deviceinfo.Names {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element)
		if err := pvcreateCmd.Run(); err != nil {
			log.WithField("error", err).Error("Failed transaction, pvcreate failed")
			return err
		}
		vgcreateCmd := exec.Command("vgcreate", strings.Replace("vg"+element, "/", "-", -1), element)
		if err := vgcreateCmd.Run(); err != nil {
			log.WithField("error", err).Error("Failed transaction, vgcreate failed")
			return err
		}
	}
	return nil
}
