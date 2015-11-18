package volume

import (
	"encoding/json"
	"path/filepath"

	"github.com/gluster/glusterd2/context"

	log "github.com/Sirupsen/logrus"
)

// AddOrUpdateVolume marshals to volume object and passes to store to add/update
func AddOrUpdateVolume(v *Volinfo) error {
	json, e := json.Marshal(v)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the volinfo object")
		return e
	}
	e = context.Store.AddOrUpdateVolumeToStore(v.Name, json)
	if e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
		return e
	}
	return nil
}

// GetVolume fetches the json object from the store and unmarshalls it into
// volinfo object
func GetVolume(name string) (*Volinfo, error) {
	var v Volinfo
	b, e := context.Store.GetVolumeFromStore(name)
	if e != nil {
		log.WithField("error", e).Error("Couldn't retrive volume from store")
		return nil, e
	}
	if e = json.Unmarshal(b, &v); e != nil {
		log.WithField("error", e).Error("Failed to unmarshal the data into volinfo object")
		return nil, e
	}
	return &v, nil
}

//DeleteVolume passes the volname to store to delete the volume object
func DeleteVolume(name string) error {
	return context.Store.DeleteVolumeFromStore(name)
}

//GetVolumes retrives the json objects from the store and converts them into
//respective volinfo objects
func GetVolumes() ([]Volinfo, error) {
	pairs, e := context.Store.GetVolumesFromStore()
	if e != nil {
		return nil, e
	}

	var vol *Volinfo
	var err error
	volumes := make([]Volinfo, len(pairs))

	for index, pair := range pairs {
		vol, err = GetVolume(filepath.Base(pair.Key))
		if err != nil || vol == nil {
			log.WithFields(log.Fields{
				"volume": pair.Key,
				"error":  err,
			}).Error("Failed to retrieve volume from the store")
			continue
		}
		volumes[index] = *vol
	}
	return volumes, nil

}

//VolumeExists check whether a given volume exist or not
func VolumeExists(name string) bool {
	return context.Store.VolumeExistsInStore(name)
}
