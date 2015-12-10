package volume

import (
	"encoding/json"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/store"
	"github.com/pborman/uuid"

	log "github.com/Sirupsen/logrus"
	etcdctx "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
)

const (
	volumePrefix string = store.GlusterPrefix + "volume/"
)

//func init() {
//context.Store.InitPrefix(volumePrefix)
//}

// AddOrUpdateVolume marshals to volume object and passes to store to add/update
func AddOrUpdateVolume(v *Volinfo) error {
	json, e := json.Marshal(v)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the volinfo object")
		return e
	}
	if _, e := context.Store.Set(etcdctx.Background(), volumePrefix+v.Name, string(json), nil); e != nil {
		log.WithField("error", e).Error("Couldn't add volume to store")
		return e
	}
	return nil
}

// GetVolume fetches the json object from the store and unmarshalls it into
// volinfo object
func GetVolume(name string) (*Volinfo, error) {
	var v Volinfo
	rsp, e := context.Store.Get(etcdctx.Background(), volumePrefix+name, nil)
	if e != nil {
		log.WithField("error", e).Error("Couldn't retrive volume from store")
		return nil, e
	}
	if e = json.Unmarshal([]byte(rsp.Node.Value), &v); e != nil {
		log.WithField("error", e).Error("Failed to unmarshal the data into volinfo object")
		return nil, e
	}
	return &v, nil
}

//DeleteVolume passes the volname to store to delete the volume object
func DeleteVolume(name string) error {
	_, err := context.Store.Delete(etcdctx.Background(), volumePrefix+name, nil)
	return err
}

func GetVolumesList() (map[string]uuid.UUID, error) {
	pairs, e := context.Store.Get(etcdctx.Background(), volumePrefix, nil)
	if e != nil || pairs == nil {
		return nil, e
	}

	volumes := make(map[string]uuid.UUID)

	for _, pair := range pairs.Node.Nodes {
		var vol Volinfo

		if err := json.Unmarshal([]byte(pair.Value), &vol); err != nil {
			log.WithFields(log.Fields{
				"volume": pair.Key,
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}

		volumes[vol.Name] = vol.ID
	}

	return volumes, nil
}

//GetVolumes retrives the json objects from the store and converts them into
//respective volinfo objects
func GetVolumes() ([]Volinfo, error) {
	pairs, e := context.Store.Get(etcdctx.Background(), volumePrefix, nil)
	if e != nil || pairs == nil {
		return nil, e
	}

	volumes := make([]Volinfo, len(pairs.Node.Nodes))

	for index, pair := range pairs.Node.Nodes {
		var vol Volinfo

		if err := json.Unmarshal([]byte(pair.Value), &vol); err != nil {
			log.WithFields(log.Fields{
				"volume": pair.Key,
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}
		volumes[index] = vol
	}

	return volumes, nil

}

//Exists check whether a given volume exist or not
func Exists(name string) bool {
	_, e := context.Store.Get(etcdctx.Background(), volumePrefix+name, nil)
	if e != nil {
		return false
	}

	return true
}
