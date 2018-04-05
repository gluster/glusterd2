package device

import (
	"encoding/json"
	"errors"

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
func CheckIfDeviceExist(reqDevices []string, metadataDevices string) ([]string, error) {

	if metadataDevices == "" {
		return reqDevices, nil
	}

	var devices []deviceapi.Info
	err := json.Unmarshal([]byte(metadataDevices), &devices)
	if err != nil {
		return nil, err
	}
	var tempDevice []string
	var flag bool
	for _, key := range reqDevices {
		flag = true
		for _, reqKey := range devices {
			if key == reqKey.Name {
				flag = false
				break
			}
		}
		if flag {
			tempDevice = append(tempDevice, key)
		}
	}
	if len(tempDevice) == 0 {
		return nil, errors.New("Devices already added")
	}
	return tempDevice, nil
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
