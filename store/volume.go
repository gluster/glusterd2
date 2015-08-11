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
			log.Error("Failed to retrieve volume %v from the store", pair.Key)
			continue
		}
		if err := json.Unmarshal(p.Value, &v); err != nil {
			log.WithField("error", err).Error("Failed to unmarshal volume %v", pair.Key)
			continue
		}
		volumes[index] = v
	}

	return volumes, nil
}

func (s *GDStore) DeleteVolume(name string) error {
	return s.Delete(volumePrefix + name)
}

func (s *GDStore) VolumeExists(name string) bool {
	if v, err := s.GetVolume(name); err != nil || v == nil {
		return false
	}
	return true
}
