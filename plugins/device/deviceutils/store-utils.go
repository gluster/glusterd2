package deviceutils

import (
	"encoding/json"
	"errors"

	peer "github.com/gluster/glusterd2/glusterd2/peer"
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
func SetDeviceState(peerID, deviceName, deviceState string) error {

	devices, err := GetDevices(peerID)
	if err != nil {
		return err
	}

	index := DeviceInList(deviceName, devices)
	if index < 0 {
		return errors.New("device does not exist in the given peer")
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
			availableSize, extentSize, err := GetVgAvailableSize(vgname)
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
