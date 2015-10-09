package store

// This file contains helper functions which make it easier to interact with
// volumes stored in the store

import (
	"github.com/docker/libkv/store"

	log "github.com/Sirupsen/logrus"
)

const (
	volumePrefix string = glusterPrefix + "volume/"
)

// AddOrUpdateVolumeToStore adds/updates given volume data in the store
func (s *GDStore) AddOrUpdateVolumeToStore(name string, b []byte) error {
	if err := s.Put(volumePrefix+name, b, nil); err != nil {
		return err
	}
	return nil
}

// GetVolumeFromStore returns the named volume from the store
func (s *GDStore) GetVolumeFromStore(name string) ([]byte, error) {
	pair, err := s.Get(volumePrefix + name)
	if err != nil || pair == nil {
		return nil, err
	}
	return pair.Value, nil
}

// GetVolumes gets all available volumes in the store
func (s *GDStore) GetVolumesFromStore() ([]*store.KVPair, error) {
	pairs, err := s.List(volumePrefix)
	if err != nil {
		log.Error("Failed to retrive volumes from the store -", err.Error())
		return nil, err
	}
	return pairs, nil
}

// DeleteVolume deletes named volume from the store
func (s *GDStore) DeleteVolumeFromStore(name string) error {
	return s.Delete(volumePrefix + name)
}

// VolumeExists checks if name volume exists in the store
func (s *GDStore) VolumeExists(name string) bool {
	if v, err := s.GetVolumeFromStore(name); err != nil || v == nil {
		return false
	}
	return true
}
