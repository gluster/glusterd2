package deviceutils

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"
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
		return nil, gderrors.ErrDeviceNotFound
	}

	if err = json.Unmarshal(resp.Kvs[0].Value, &device); err != nil {
		return nil, err
	}

	return &device, nil
}

// SetDeviceState sets device state and updates device state in etcd
func SetDeviceState(peerID, deviceName, deviceState string) error {
	dev, err := GetDevice(peerID, deviceName)
	if err != nil {
		return err
	}

	dev.State = deviceState
	return AddOrUpdateDevice(*dev)
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
func UpdateDeviceFreeSize(peerID, device string) error {
	dev, err := GetDevice(peerID, device)
	if err != nil {
		return err
	}

	availableSize, extentSize, err := lvmutils.GetVgAvailableSize(dev.VgName())
	if err != nil {
		return err
	}
	dev.AvailableSize = availableSize
	dev.UsedSize = dev.TotalSize - availableSize
	dev.ExtentSize = extentSize
	return AddOrUpdateDevice(*dev)
}

// UpdateDeviceFreeSizeByVg updates the actual available size of VG
func UpdateDeviceFreeSizeByVg(peerID, vgname string) error {
	devs, err := GetDevices(peerID)
	if err != nil {
		return err
	}

	for _, dev := range devs {
		if dev.VgName() == vgname {
			availableSize, extentSize, err := lvmutils.GetVgAvailableSize(vgname)
			if err != nil {
				return err
			}
			dev.AvailableSize = availableSize
			dev.UsedSize = dev.TotalSize - availableSize
			dev.ExtentSize = extentSize
			return AddOrUpdateDevice(dev)
		}
	}
	return gderrors.ErrDeviceNotFound
}

//IsVgExist checks whether the given vg exist in the device list for the local peer
func IsVgExist(vgname string) bool {
	peerID := gdctx.MyUUID.String()
	deviceDetails, err := GetDevices(peerID)
	if err != nil {
		return false
	}

	for _, dev := range deviceDetails {
		if dev.VgName() == vgname {
			return true
		}
	}
	return false
}

// GetDeviceAvailableSize gets the device size and vgName using device Path
func GetDeviceAvailableSize(peerID, device string) (uint64, error) {
	dev, err := GetDevice(peerID, device)
	if err != nil {
		return 0, err
	}
	return dev.AvailableSize, nil
}

// CheckForAvailableVgSize prepares a brickName to vgName mapping in order to use while expanding lvm.
// Also it checks for sufficient space available on devices of current node.
func CheckForAvailableVgSize(expansionSize uint64, bricksInfo []brick.Brickinfo) (map[string]string, bool, error) {

	// map of brickName to vgName. Used while extending lv
	brickVgMapping := make(map[string]string)

	// map of deviceName to required free size in that device
	requiredDeviceSizeMap := make(map[string]uint64)

	// Map deviceName to required free space on the device
	for _, b := range bricksInfo {
		if _, ok := requiredDeviceSizeMap[b.MountInfo.DevicePath]; !ok {
			requiredDeviceSizeMap[b.MountInfo.DevicePath] = expansionSize
		} else {
			requiredDeviceSizeMap[b.MountInfo.DevicePath] += expansionSize
		}
	}

	// Check in the map prepared in last step by looking through devices names and device available size of bricks from current node.
	for _, b := range bricksInfo {
		// retrieve device available size by device Name and return the vgName and available device Size.
		deviceSize, err := GetDeviceAvailableSize(b.PeerID.String(), b.RootDevice)
		if err != nil {
			return map[string]string{}, false, err
		}
		if requiredDeviceSizeMap[b.MountInfo.DevicePath] > deviceSize {
			return map[string]string{}, false, nil
		}
		brickVgMapping[b.Path] = b.DeviceInfo.VgName
	}

	return brickVgMapping, true, nil
}
