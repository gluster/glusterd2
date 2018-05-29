package deviceutils

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
	var deviceInfo []deviceapi.Info
	if len(peerInfo.Metadata["_devices"]) > 0 {
		if err := json.Unmarshal([]byte(peerInfo.Metadata["_devices"]), &deviceInfo); err != nil {
			return nil, err
		}
	}
	return deviceInfo, nil
}

// DeviceExist checks the given device existence
func DeviceExist(reqDevice string, devices []deviceapi.Info) bool {
	for _, key := range devices {
		if reqDevice == key.Name {
			return true
		}
	}
	return false
}

// AddDevice adds device to peerinfo
func AddDevice(device deviceapi.Info, peerID string) error {
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
