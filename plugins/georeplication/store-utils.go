package georeplication

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/store"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	log "github.com/sirupsen/logrus"
)

const (
	georepPrefix string = store.GlusterPrefix + "georeplication/"
)

// getSession fetches the json object from the store and unmarshalls it into
// Info object
func getSession(masterid string, slaveid string) (*georepapi.GeorepSession, error) {
	var v georepapi.GeorepSession
	resp, e := store.Store.Get(context.TODO(), georepPrefix+masterid+"/"+slaveid)
	if e != nil {
		log.WithError(e).Error("Couldn't retrive geo-replication session from store")
		return nil, e
	}

	if resp.Count != 1 {
		return nil, &ErrGeorepSessionNotFound{}
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &v); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into georepinfo object")
		return nil, e
	}
	return &v, nil
}

// addOrUpdateSession marshals the georep session object and passes to store to add/update
func addOrUpdateSession(v *georepapi.GeorepSession) error {
	json, e := json.Marshal(v)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the Info object")
		return e
	}

	_, e = store.Store.Put(context.TODO(), georepPrefix+v.MasterID.String()+"/"+v.SlaveID.String(), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add georeplication session to store")
		return e
	}
	return nil
}

// deleteSession deletes the georep session object from store
func deleteSession(mastervolid string, slavevolid string) error {
	_, e := store.Store.Delete(context.TODO(), georepPrefix+mastervolid+"/"+slavevolid)
	if e != nil {
		log.WithError(e).Error("Couldn't delete georeplication session from store")
		return e
	}
	return nil
}
