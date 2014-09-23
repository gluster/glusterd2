package consul

import (
	"encoding/json"

	"github.com/armon/consul-api"
	"github.com/kshlm/glusterd2/volume"
)

const (
	volumePrefix string = glusterPrefix + "volume/"
)

func (c *Consul) AddVolume(v *volume.Volinfo) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}

	pair := consulapi.KVPair{Key: volumePrefix + v.Name, Value: json}

	if _, err := c.kv.Put(&pair, nil); err != nil {
		return err
	}

	return nil
}

func (c *Consul) GetVolume(name string) (*volume.Volinfo, error) {
	pair, _, err := c.kv.Get(volumePrefix+name, nil)
	if err != nil || pair == nil {
		return nil, err
	}

	var v volume.Volinfo
	if err := json.Unmarshal(pair.Value, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Consul) VolumeExists(name string) bool {
	if v, err := c.GetVolume(name); err != nil || v == nil {
		return false
	}
	return true
}
