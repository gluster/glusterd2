package heketi

import (
	"fmt"
	"os/exec"

	heketi "github.com/gluster/glusterd2/plugins/heketi/api"
	"github.com/gluster/glusterd2/transaction"
)

func txnHeketiPrepareDevice(c transaction.TxnCtx) error {
	var deviceinfo heketi.DeviceInfo

	if err := c.Get("nodeid", &deviceinfo.NodeID); err != nil {
		return err
	}
	if err := c.Get("devicename", &deviceinfo.DeviceName); err != nil {
		return err
	}


	cmd1 := exec.Command("pvcreate", "--metadatasize=128M", "--dataalignment=256K", "/dev/"+deviceinfo.DeviceName)
	if err := cmd1.Run(); err != nil {
		fmt.Printf("failed to run pvcreate")
	}
	cmd2 := exec.Command("vgcreate", deviceinfo.DeviceName, "/dev/"+deviceinfo.DeviceName)
	if err := cmd2.Run(); err != nil {
		fmt.Printf("failed to run vgcreate")
	}

	return nil
}

func txnHeketiCreateBrick(c transaction.TxnCtx) error {

	return nil
}
