package devicecommands

import (
	"os/exec"

	device "github.com/gluster/glusterd2/glusterd2/device"
	"github.com/gluster/glusterd2/glusterd2/transaction"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var deviceinfo device.Info
	if err := c.Get("nodeid", &deviceinfo.NodeID); err != nil {
		return err
	}
	if err := c.Get("devicename", &deviceinfo.DeviceName); err != nil {
		return err
	}
	for _, element := range deviceinfo.DeviceName {
		pvcreateCmd := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", element)
		if err := pvcreateCmd.Run(); err != nil {
			return err
		}
		vgcreateCmd := exec.Command("vgcreate", "vg"+element, element)
		if err := vgcreateCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
