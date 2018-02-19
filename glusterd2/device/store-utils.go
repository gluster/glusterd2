// Package device stores device information in the store
package device

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
)

const (
	devicePrefix string = "devices/"
	peerPrefix string = "peers/"
)

// GetDevice returns devices of specified peer from the store
func GetDevice(peerid string) (*api.Device, error) {
	resp, err := store.Store.Get(context.TODO(), peerPrefix+peerid+devicePrefix)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Printing Get Device %s", resp)
	/*if len(resp.Kvs) > 0 {
		var deviceDetail api.Device
		if err := json.Unmarshal(resp.Kvs[0].Value, &deviceDetail); err != nil {
			return nil, err
		}
		return &deviceDetail, nil
	}*/

	return nil, nil
}

// AddOrUpdateDevice adds device to specific peer
func AddOrUpdateDevice(d *api.Device) error {
	json, err := json.Marshal(d)
	if err != nil {
		return err
	}

	idStr := d.PeerID.String()

	if _, err := store.Store.Put(context.TODO(), peerPrefix+idStr+devicePrefix, string(json)); err != nil {
		return err
	}

	return nil
}
