package deviceutils

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	peer "github.com/gluster/glusterd2/glusterd2/peer"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// GetDevices returns devices of specified peer/peers from the store
// if no peers are specified, it returns devices of all peers
func GetDevices(peerIds ...string) ([]deviceapi.Info, error) {

	var peers []*peer.Peer
	var err error
	if len(peerIds) > 0 {
		for _, peerID := range peerIds {
			var peerInfo *peer.Peer
			peerInfo, err = peer.GetPeer(peerID)
			if err != nil {
				return nil, err
			}
			peers = append(peers, peerInfo)
		}
	} else {
		peers, err = peer.GetPeers()
		if err != nil {
			return nil, err
		}
	}
	var devices []deviceapi.Info
	for _, peerInfo := range peers {
		deviceInfo, err := GetDevicesFromPeer(peerInfo)
		if err != nil {
			return nil, err
		}
		devices = append(devices, deviceInfo...)
	}
	return devices, nil
}

// GetDevicesFromPeer returns devices from peer object.
func GetDevicesFromPeer(peerInfo *peer.Peer) ([]deviceapi.Info, error) {

	var deviceInfo []deviceapi.Info
	if _, exists := peerInfo.Metadata["_devices"]; exists {
		if err := json.Unmarshal([]byte(peerInfo.Metadata["_devices"]), &deviceInfo); err != nil {
			return nil, err
		}
	}
	return deviceInfo, nil
}

// SetDeviceState sets device state and updates device state in etcd
func SetDeviceState(peerID, device, deviceState string) error {

	devices, err := GetDevices(peerID)
	if err != nil {
		return err
	}

	index := DeviceInList(device, devices)
	if index < 0 {
		return errors.New("device does not exist in the given peer")
	}
	if devices[index].State == deviceState {
		return nil
	}
	devices[index].State = deviceState
	return updateDevices(peerID, devices)
}

func updateDevices(peerID string, devices []deviceapi.Info) error {
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return err
	}
	deviceJSON, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	peerInfo.Metadata["_devices"] = string(deviceJSON)
	return peer.AddOrUpdatePeer(peerInfo)
}

// DeviceInList returns index of device if device is present in list else returns -1.
func DeviceInList(reqDevice string, devices []deviceapi.Info) int {
	for index, key := range devices {
		if reqDevice == key.Name {
			return index
		}
	}
	return -1
}

// AddDevice adds device to peerinfo
func AddDevice(device deviceapi.Info) error {
	deviceDetails, err := GetDevices(device.PeerID)
	if err != nil {
		return err
	}
	peerInfo, err := peer.GetPeer(device.PeerID)
	if err != nil {
		return err
	}
	var devices []deviceapi.Info
	if deviceDetails != nil {
		devices = append(deviceDetails, device)
	} else {
		devices = append(devices, device)
	}

	deviceJSON, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	peerInfo.Metadata["_devices"] = string(deviceJSON)
	err = peer.AddOrUpdatePeer(peerInfo)
	if err != nil {
		return err
	}

	return nil
}

// UpdateDeviceFreeSize updates the actual available size of VG
func UpdateDeviceFreeSize(peerid, vgname string) error {
	deviceDetails, err := GetDevices(peerid)
	if err != nil {
		return err
	}

	peerInfo, err := peer.GetPeer(peerid)
	if err != nil {
		return err
	}

	for idx, dev := range deviceDetails {
		if dev.VgName == vgname {
			availableSize, extentSize, err := lvmutils.GetVgAvailableSize(vgname)
			if err != nil {
				return err
			}
			deviceDetails[idx].AvailableSize = availableSize
			deviceDetails[idx].ExtentSize = extentSize
		}
	}

	deviceJSON, err := json.Marshal(deviceDetails)
	if err != nil {
		return err
	}

	peerInfo.Metadata["_devices"] = string(deviceJSON)
	err = peer.AddOrUpdatePeer(peerInfo)
	if err != nil {
		return err
	}

	return nil
}

//IsVgExist checks whether the given vg exist in the device list for the local peer
func IsVgExist(vgname string) bool {
	peerid := gdctx.MyUUID.String()
	deviceDetails, err := GetDevices(peerid)
	if err != nil {
		return false
	}

	for _, dev := range deviceDetails {
		if dev.VgName == vgname {
			return true
		}
	}
	return false
}

// getDeviceAvailableSize gets the device size and vgName using device Path
func getDeviceAvailableSize(peerid, devicePath string) (string, uint64, error) {
	devices, err := GetDevices(peerid)
	if err != nil {
		return "", 0, err
	}
	deviceVgName := strings.Split(devicePath, "/")[2]
	for _, d := range devices {
		if d.VgName == deviceVgName {
			return d.VgName, d.AvailableSize, nil
		}
	}
	return "", 0, gderrors.ErrDeviceNameNotFound
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
		vgName, deviceSize, err := getDeviceAvailableSize(b.PeerID.String(), b.MountInfo.DevicePath)
		if err != nil {
			return map[string]string{}, false, err
		}
		if requiredDeviceSizeMap[b.MountInfo.DevicePath] > deviceSize {
			return map[string]string{}, false, nil
		}
		brickVgMapping[b.Path] = vgName
	}

	return brickVgMapping, true, nil
}
