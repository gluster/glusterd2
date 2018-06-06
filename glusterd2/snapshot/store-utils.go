package snapshot

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/coreos/etcd/clientv3"
	gdstore "github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	log "github.com/sirupsen/logrus"
)

const (
	snapPrefix string = "snaps/"
)

var (
	//ExistsFunc check whether a given snapshot exist or not
	ExistsFunc = Exists
	// AddOrUpdateSnapFunc marshals to snap object and passes to store to add/update
	AddOrUpdateSnapFunc = AddOrUpdateSnap
)

//Exists check whether a given snapshot exist or not
func Exists(name string) bool {
	resp, e := gdstore.Get(context.TODO(), snapPrefix+name)
	if e != nil {
		return false
	}

	return resp.Count == 1
}

//GetSnapshots retrives the json objects from the store and converts them into
//respective volinfo objects
func GetSnapshots() ([]*Snapinfo, error) {
	resp, e := gdstore.Get(context.TODO(), snapPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	snaps := make([]*Snapinfo, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var snap Snapinfo

		if err := json.Unmarshal(kv.Value, &snap); err != nil {
			log.WithFields(log.Fields{
				"volume": string(kv.Key),
				"error":  err,
			}).Error("Failed to unmarshal volume")
			continue
		}

		snaps[i] = &snap
	}

	return snaps, nil
}

//GetSnapshotVolumes return the volfile for all snapshots
func GetSnapshotVolumes() ([]*volume.Volinfo, error) {
	var vols []*volume.Volinfo
	resp, err := GetSnapshots()
	if err != nil {
		return vols, err
	}
	for _, snap := range resp {
		vols = append(vols, &snap.SnapVolinfo)
	}
	return vols, nil
}

// AddOrUpdateSnap marshals to volume object and passes to store to add/update
func AddOrUpdateSnap(snapInfo *Snapinfo) error {
	json, e := json.Marshal(snapInfo)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the volinfo object")
		return e
	}

	_, e = gdstore.Store.Put(context.TODO(), GetStorePath(snapInfo), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add volume to store")
		return e
	}
	return nil
}

// GetSnapshot fetches the json object from the store and unmarshalls it into
// Snapinfo object
func GetSnapshot(name string) (*Snapinfo, error) {
	var snap Snapinfo
	resp, e := gdstore.Store.Get(context.TODO(), snapPrefix+name)
	if e != nil {
		log.WithError(e).Error("Couldn't retrive volume from store")
		return nil, e
	}

	if resp.Count != 1 {
		log.WithField("volume", name).Error("volume not found")
		return nil, errors.New("volume not found")
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &snap); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into volinfo object")
		return nil, e
	}
	return &snap, nil
}

//DeleteSnapshot passes the snap path to store to delete the snap object
func DeleteSnapshot(snapInfo *Snapinfo) error {
	_, e := gdstore.Store.Delete(context.TODO(), GetStorePath(snapInfo))
	return e
}

//GetStorePath return snapshot path for etcd store
func GetStorePath(snapInfo *Snapinfo) string {
	return snapPrefix + snapInfo.SnapVolinfo.Name
}
