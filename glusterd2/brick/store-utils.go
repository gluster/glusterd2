package brick

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/store"

	log "github.com/sirupsen/logrus"

	"github.com/coreos/etcd/clientv3"
)

const (
	// glusterfsdPrefix represents the etcd end-point for all brick processes
	glusterfsdPrefix string = "glusterfsds/"
)

// UpdateBrickProcess adds a new brick process to store or updates existing brick process
func UpdateBrickProcess(bp *Glusterfsd) error {
	json, e := json.Marshal(bp)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the glusterfsd object")
		return e
	}

	_, e = store.Store.Put(context.TODO(), glusterfsdPrefix+bp.ID(), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add glusterfsd to store")
		return e
	}
	log.WithField("brick", bp.Binfo.Path).Info("Updated brick process")
	return nil
}

// DeleteBrickProcess removes brick process instance from store
func DeleteBrickProcess(bp *Glusterfsd) error {
	_, err := store.Store.Delete(context.TODO(), glusterfsdPrefix+bp.ID())

	return err
}

// GetBrickProcessByPort fetches the json object from the store and unmarshalls it into
// volinfo object
func GetBrickProcessByPort(p int) (*Glusterfsd, error) {
	bps, err := GetBrickProcesses()
	if err != nil {
		return nil, err
	}

	for _, bp := range bps {
		if bp.Port == p {
			return bp, nil
		}
	}

	return nil, errors.New("Didn't find brick process for port")
}

//GetBrickProcesses retrives the json objects from the store and converts them into
//respective BrickProcess objects
func GetBrickProcesses() ([]*Glusterfsd, error) {
	resp, e := store.Store.Get(context.TODO(), glusterfsdPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	bps := make([]*Glusterfsd, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var bp Glusterfsd

		if err := json.Unmarshal(kv.Value, &bp); err != nil {
			log.WithFields(log.Fields{
				"brickprocess": string(kv.Key),
				"error":        err,
			}).Error("Failed to unmarshal glusterfsd info")
			continue
		}

		log.WithField("brick", bp.Binfo.Path).Info("Got brick process from etcd store")
		bps[i] = &bp
	}

	return bps, nil
}
