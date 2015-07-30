package store

// This file contains helper functions which make it easier to interact with
// volumes stored in the store

import (
	"encoding/json"

	"github.com/kshlm/glusterd2/volume"
)

const (
	volumePrefix string = glusterPrefix + "volume/"
)

func (s *GDStore) AddVolume(v *volume.Volinfo) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if err := s.Put(volumePrefix+v.Name, json, nil); err != nil {
		return err
	}

	return nil
}

func (s *GDStore) GetVolume(name string) (*volume.Volinfo, error) {
	pair, err := s.Get(volumePrefix + name)
	if err != nil || pair == nil {
		return nil, err
	}

	var v volume.Volinfo
	if err := json.Unmarshal(pair.Value, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *GDStore) VolumeExists(name string) bool {
	if v, err := s.GetVolume(name); err != nil || v == nil {
		return false
	}
	return true
}
