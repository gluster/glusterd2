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
	if len(peerInfo.Metadata["_devices"]) > 0 {
		var deviceInfo []deviceapi.Info
		if err := json.Unmarshal([]byte(peerInfo.Metadata["_devices"]), &deviceInfo); err != nil {
			return nil, err
		}
		return deviceInfo, nil
	}
	return nil, nil
}

func checkIfDeviceExist(reqDevice string, devices []deviceapi.Info) bool {
	for _, key := range devices {
		if reqDevice == key.Name {
			return true
		}
	}
	return false
}

func addDevice(device deviceapi.Info, peerID string) error {
	deviceDetails, err := GetDevices(peerID)
	if err != nil {
		return err
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return err
	}
	var devices []deviceapi.Info
	if deviceDetails != nil {
		devices = append(deviceDetails, device)
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
