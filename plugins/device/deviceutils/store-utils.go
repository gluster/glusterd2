package deviceutils

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/store"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/coreos/etcd/clientv3"
)

const (
	devicePrefix string = "devices/"
)

// GetDevices returns devices of specified peer/peers from the store
// if no peers are specified, it returns devices of all peers
func GetDevices(peerIds ...string) ([]deviceapi.Info, error) {
	var devices []deviceapi.Info
	var err error
	var resp *clientv3.GetResponse

	if len(peerIds) > 0 {
		for _, peerID := range peerIds {
			resp, err = store.Get(context.TODO(), devicePrefix+peerID+"/", clientv3.WithPrefix())
			if err != nil {
				return nil, err
			}
		}
	} else {
		resp, err = store.Get(context.TODO(), devicePrefix, clientv3.WithPrefix())
		if err != nil {
			return nil, err
		}
	}
	for _, kv := range resp.Kvs {
		var dev deviceapi.Info

		if err = json.Unmarshal(kv.Value, &dev); err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}
	return devices, nil
}

// GetDevice returns device of specified peer and device name
func GetDevice(peerID, deviceName string) (*deviceapi.Info, error) {
	var device deviceapi.Info
	var err error

	resp, err := store.Get(context.TODO(), devicePrefix+peerID+"/"+deviceName)
	if err != nil {
		return nil, err
	}

	if resp.Count != 1 {
		return nil, ErrDeviceNotFound
	}

	if err = json.Unmarshal(resp.Kvs[0].Value, &device); err != nil {
		return nil, err
	}

	return &device, nil
}

// SetDeviceState sets device state and updates device state in etcd
func SetDeviceState(peerID, deviceName, deviceState string) error {
	resp, err := store.Get(context.TODO(), devicePrefix+peerID+"/"+deviceName)
	if err != nil {
		return err
	}

	if resp.Count != 1 {
		return ErrDeviceNotFound
	}

	var dev deviceapi.Info
	if err := json.Unmarshal(resp.Kvs[0].Value, &dev); err != nil {
		return err
	}

	dev.State = deviceState
	return AddOrUpdateDevice(dev)
}

// AddOrUpdateDevice adds device to peerinfo
func AddOrUpdateDevice(device deviceapi.Info) error {
	json, err := json.Marshal(device)
	if err != nil {
		return err
	}

	storeKey := device.PeerID.String() + "/" + device.Device

	if _, err := store.Put(context.TODO(), devicePrefix+storeKey, string(json)); err != nil {
		return err
	}

	return nil
}

// UpdateDeviceFreeSize updates the actual available size of VG
func UpdateDeviceFreeSize(peerid, device string, size uint64, extentSize uint64) error {
	deviceDetails, err := GetDevice(peerid, device)
	if err != nil {
		return err
	}
	deviceDetails.AvailableSize = size
	deviceDetails.ExtentSize = extentSize
	return AddOrUpdateDevice(*deviceDetails)
}
