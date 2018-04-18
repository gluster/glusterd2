package device

import (
	"encoding/json"

	peer "github.com/gluster/glusterd2/glusterd2/peer"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// GetDevices returns devices of specified peer from the store
func GetDevices(peerID string) ([]deviceapi.Info, error) {
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return nil, err
	}
	if len(peerInfo.Metadata["devices"]) > 0 {
		var deviceInfo []deviceapi.Info
		if err := json.Unmarshal([]byte(peerInfo.Metadata["devices"]), &deviceInfo); err != nil {
			return nil, err
		}
		return deviceInfo, nil
	}
	return nil, nil
}

//CheckIfDeviceExist returns error if all devices already exist or returns list of devices to be added
func CheckIfDeviceExist(reqDevices []string, devices []deviceapi.Info) bool {

	for _, key := range reqDevices {
		for _, reqKey := range devices {
			if key == reqKey.Name {
				return false
			}
		}
	}
	return true
}

// AddDevices adds device to specific peer
func AddDevices(devices []deviceapi.Info, peerID string) error {
	deviceDetails, err := GetDevices(peerID)
	if err != nil {
		return err
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return err
	}
	if deviceDetails != nil {
		devices = append(devices, deviceDetails...)
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
