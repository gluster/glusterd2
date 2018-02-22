// Package device stores device information in the store
package device

import (
	"encoding/json"

	peer "github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
)

// GetDevices returns devices of specified peer from the store
func GetDevices(peerID string) ([]api.DeviceInfo, error) {
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return nil, err
	}
	if len(peerInfo.MetaData["devices"]) > 0 {
		var deviceInfo []api.DeviceInfo
		if err := json.Unmarshal([]byte(peerInfo.MetaData["devices"]), &deviceInfo); err != nil {
			return nil, err
		}
		return deviceInfo, nil
	}
	return nil, nil
}

// AddDevices adds device to specific peer
func AddDevices(devices []api.DeviceInfo, peerID string) error {
	deviceDetails, err := GetDevices(peerID)
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		return err
	}
	if deviceDetails != nil {
		for _, element := range devices {
			devices = append(deviceDetails, element)
		}
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
