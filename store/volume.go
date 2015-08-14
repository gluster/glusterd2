package store

// This file contains helper functions which make it easier to interact with
// volumes stored in the store

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/kshlm/glusterd2/volume"
)

const (
	volumePrefix string = glusterPrefix + "volume/"
)

// AddOrUpdateVolume adds/updates given volume in the store
func (s *GDStore) AddOrUpdateVolume(v *volume.Volinfo) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if err := s.Put(volumePrefix+v.Name, json, nil); err != nil {
		return err
	}

	return nil
}

// GetVolume returns the named volume from the store
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

// GetVolumes gets all available volumes in the store
func (s *GDStore) GetVolumes() ([]volume.Volinfo, error) {
	pairs, err := s.List(volumePrefix)
	if err != nil || pairs == nil {
		return nil, err
	}

	var v volume.Volinfo
	volumes := make([]volume.Volinfo, len(pairs))
	for index, pair := range pairs {
		p, err := s.Get(pair.Key)
		if err != nil || p == nil {
			log.WithFields(log.Fields{
				"volume": pair.Key,
				"error":  err,
			}).Error("Failed to retrieve volume from the store")
			continue
		}
		if err := json.Unmarshal(p.Value, &v); err != nil {
			log.WithFields(log.Fields{
				"volume": pair.Key,
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}
		volumes[index] = v
	}

	return volumes, nil
}

// DeleteVolume deletes named volume from the store
func (s *GDStore) DeleteVolume(name string) error {
	return s.Delete(volumePrefix + name)
}

// VolumeExists checks if name volume exists in the store
func (s *GDStore) VolumeExists(name string) bool {
	if v, err := s.GetVolume(name); err != nil || v == nil {
		return false
	}
	return true
}
