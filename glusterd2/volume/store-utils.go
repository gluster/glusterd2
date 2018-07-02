package volume

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/store"
	gderror "github.com/gluster/glusterd2/pkg/errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	volumePrefix string = "volumes/"
)

// metadataFilter is a filter type
type metadataFilter uint32

// GetVolumes Filter Types
const (
	noKeyAndValue metadataFilter = iota
	onlyKey
	onlyValue
	keyAndValue
)

var (
	// AddOrUpdateVolumeFunc marshals to volume object and passes to store to add/update
	AddOrUpdateVolumeFunc = AddOrUpdateVolume
)

// AddOrUpdateVolume marshals to volume object and passes to store to add/update
func AddOrUpdateVolume(v *Volinfo) error {
	json, e := json.Marshal(v)
	if e != nil {
		log.WithError(e).Error("Failed to marshal the volinfo object")
		return e
	}

	_, e = store.Put(context.TODO(), volumePrefix+v.Name, string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add volume to store")
		return e
	}
	return nil
}

// GetVolume fetches the json object from the store and unmarshalls it into
// volinfo object
func GetVolume(name string) (*Volinfo, error) {
	var v Volinfo
	resp, e := store.Get(context.TODO(), volumePrefix+name)
	if e != nil {
		log.WithError(e).Error("Couldn't retrive volume from store")
		return nil, e
	}

	if resp.Count != 1 {
		log.WithField("volume", name).Error("volume not found")
		return nil, gderror.ErrVolNotFound
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &v); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into volinfo object")
		return nil, e
	}
	return &v, nil
}

//DeleteVolume passes the volname to store to delete the volume object
func DeleteVolume(name string) error {
	_, e := store.Delete(context.TODO(), volumePrefix+name)
	return e
}

// GetVolumesList returns a map of volume names to their UUIDs
func GetVolumesList() (map[string]uuid.UUID, error) {
	resp, e := store.Get(context.TODO(), volumePrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	volumes := make(map[string]uuid.UUID)

	for _, kv := range resp.Kvs {
		var vol Volinfo

		if err := json.Unmarshal(kv.Value, &vol); err != nil {
			log.WithFields(log.Fields{
				"volume": string(kv.Key),
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}

		volumes[vol.Name] = vol.ID
	}

	return volumes, nil
}

// getFilterType return the filter type for volume list/info
func getFilterType(filterParams map[string]string) metadataFilter {
	_, key := filterParams["key"]
	_, value := filterParams["value"]
	if key && !value {
		return onlyKey
	} else if value && !key {
		return onlyValue
	} else if value && key {
		return keyAndValue
	}
	return noKeyAndValue
}

//GetVolumes retrives the json objects from the store and converts them into
//respective volinfo objects
func GetVolumes(filterParams ...map[string]string) ([]*Volinfo, error) {
	resp, e := store.Get(context.TODO(), volumePrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	var filterType metadataFilter
	if len(filterParams) == 0 {
		filterType = noKeyAndValue
	} else {
		filterType = getFilterType(filterParams[0])
	}

	var volumes []*Volinfo

	for _, kv := range resp.Kvs {
		var vol Volinfo

		if err := json.Unmarshal(kv.Value, &vol); err != nil {
			log.WithFields(log.Fields{
				"volume": string(kv.Key),
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}
		switch filterType {

		case onlyKey:
			if _, keyFound := vol.Metadata[filterParams[0]["key"]]; keyFound {
				volumes = append(volumes, &vol)
			}
		case onlyValue:
			for _, value := range vol.Metadata {
				if value == filterParams[0]["value"] {
					volumes = append(volumes, &vol)
				}
			}
		case keyAndValue:
			if value, keyFound := vol.Metadata[filterParams[0]["key"]]; keyFound {
				if value == filterParams[0]["value"] {
					volumes = append(volumes, &vol)
				}
			}
		default:
			volumes = append(volumes, &vol)

		}
	}

	return volumes, nil
}

// GetAllBricksInCluster returns all bricks in the cluster. These bricks
// belong to different volumes.
func GetAllBricksInCluster() ([]brick.Brickinfo, error) {

	volumes, err := GetVolumes()
	if err != nil {
		return nil, err
	}

	var bricks []brick.Brickinfo
	for _, volinfo := range volumes {
		bricks = append(bricks, volinfo.GetBricks()...)
	}

	return bricks, nil
}

// AreReplicateVolumesRunning retrieves the volinfo objects from GetVolumes() function
// and checks if all replicate, Disperse volumes are stopped before
// stopping the self heal daemon.
func AreReplicateVolumesRunning() (bool, error) {
	volumes, e := GetVolumes()
	if e != nil {
		return false, e
	}
	for _, v := range volumes {
		if (v.Type == Replicate || v.Type == Disperse || v.Type == DistReplicate || v.Type == DistDisperse) && v.State == VolStarted {
			return true, nil
		}
	}

	return false, nil
}

//Exists check whether a given volume exist or not
func Exists(name string) bool {
	resp, e := store.Get(context.TODO(), volumePrefix+name)
	if e != nil {
		return false
	}

	return resp.Count == 1
}
