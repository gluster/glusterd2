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
	if len(peerInfo.MetaData["devices"]) > 0 {
		var deviceInfo []deviceapi.Info
		if err := json.Unmarshal([]byte(peerInfo.MetaData["devices"]), &deviceInfo); err != nil {
			return nil, err
		}
		return deviceInfo, nil
	}
	return nil, nil
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
	peerInfo.MetaData["devices"] = string(deviceJSON)
	err = peer.AddOrUpdatePeer(peerInfo)
	if err != nil {
		return err
	}

	return nil

}
