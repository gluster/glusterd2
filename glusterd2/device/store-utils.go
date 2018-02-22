// Package device stores device information in the store
package device

import (
	"encoding/json"

	peer "github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
)

const (
	devicePrefix string = "/devices/"
	peerPrefix   string = "peers/"
)

// GetDevice returns devices of specified peer from the store
func GetDevice(peerID string) ([]api.DeviceInfo, error) {
	p, err := peer.GetPeer(peerID)
	if err != nil {
		return nil, err
	}
	if len(p.MetaData["devices"]) > 0 {
		var deviceInfo []api.DeviceInfo
		if err := json.Unmarshal([]byte(p.MetaData["devices"]), &deviceInfo); err != nil {
			return nil, err
		}
		return deviceInfo, nil
	}
	return nil, nil
}

// AddOrUpdateDevice adds device to specific peer
func AddOrUpdateDevice(d []api.DeviceInfo, peerID string) error {
	deviceJSON, err := json.Marshal(d)
	if err != nil {
		return err
	}

	p, err := peer.GetPeer(peerID)
	p.MetaData["devices"] = string(deviceJSON)
	err = peer.AddOrUpdatePeer(p)
	if err != nil {
		return err
	}

	return nil

}
