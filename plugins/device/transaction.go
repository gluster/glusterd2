package device

import (
	"github.com/gluster/glusterd2/glusterd2/provisioners"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/pborman/uuid"
)

func txnPrepareDevice(c transaction.TxnCtx) error {
	var peerID string
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerid").Error("Failed to get key from transaction context")
		return err
	}

	var device string
	if err := c.Get("device", &device); err != nil {
		c.Logger().WithError(err).WithField("key", "device").Error("Failed to get key from transaction context")
		return err
	}

	var provisionerType string
	if err := c.Get("provisioner", &provisionerType); err != nil {
		c.Logger().WithError(err).WithField("key", "provisioner").Error("Failed to get key from transaction context")
		return err
	}

	var provisioner provisioners.Provisioner
	var err error
	if provisionerType == "" {
		provisioner = provisioners.GetDefault()
	} else {
		provisioner, err = provisioners.Get(provisionerType)
		if err != nil {
			c.Logger().WithError(err).WithField("name", provisionerType).Error("invalid provisioner")
			return err
		}
	}

	var deviceInfo deviceapi.Info

	err = provisioner.Register(device)
	if err != nil {
		c.Logger().WithError(err).WithField("device", device).Error("failed to register device")
		return err
	}

	c.Logger().WithField("device", device).Info("Device setup successful, setting device status to 'Enabled'")

	availableSize, extentSize, err := provisioner.AvailableSize(device)
	if err != nil {
		return err
	}
	deviceInfo = deviceapi.Info{
		Device:        device,
		State:         deviceapi.DeviceEnabled,
		AvailableSize: availableSize,
		ExtentSize:    extentSize,
		PeerID:        uuid.Parse(peerID),
	}

	err = deviceutils.AddOrUpdateDevice(deviceInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Couldn't add deviceinfo to store")
		return err
	}
	return nil
}
