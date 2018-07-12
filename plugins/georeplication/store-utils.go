package georeplication

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/store"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

const (
	georepPrefix        string = "georeplication/"
	georepSSHKeysPrefix string = "georeplication-ssh-keys/"
)

// getSession fetches the json object from the store and unmarshalls it into
// Info object
func getSession(masterid string, remoteid string) (*georepapi.GeorepSession, error) {
	var v georepapi.GeorepSession
	resp, e := store.Get(context.TODO(), georepPrefix+masterid+"/"+remoteid)
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
		log.WithError(e).Error("Failed to marshal the Info object")
		return e
	}

	_, e = store.Put(context.TODO(), georepPrefix+v.MasterID.String()+"/"+v.RemoteID.String(), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add georeplication session to store")
		return e
	}
	return nil
}

// deleteSession deletes the georep session object from store
func deleteSession(mastervolid string, remotevolid string) error {
	_, e := store.Delete(context.TODO(), georepPrefix+mastervolid+"/"+remotevolid)
	if e != nil {
		log.WithError(e).Error("Couldn't delete georeplication session from store")
		return e
	}
	return nil
}

// getSessionList gets list of Geo-replication sessions
func getSessionList() ([]*georepapi.GeorepSession, error) {
	resp, e := store.Get(context.TODO(), georepPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	sessions := make([]*georepapi.GeorepSession, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var session georepapi.GeorepSession

		if err := json.Unmarshal(kv.Value, &session); err != nil {
			log.WithError(err).WithField("session", string(kv.Key)).Error("Failed to unmarshal Geo-replication session")
			continue
		}

		sessions[i] = &session
	}

	return sessions, nil
}

// addOrUpdateSSHKeys marshals the georep SSH Public keys to add/update
func addOrUpdateSSHKey(volname string, sshkey georepapi.GeorepSSHPublicKey) error {
	json, e := json.Marshal(sshkey)
	if e != nil {
		log.WithError(e).Error("Failed to marshal the sshkeys object")
		return e
	}

	_, e = store.Put(context.TODO(), georepSSHKeysPrefix+volname+"/"+sshkey.PeerID.String(), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add SSH public key to Store")
		return e
	}
	return nil
}

// getSSHPublicKeys returns list of SSH public keys
func getSSHPublicKeys(volname string) ([]georepapi.GeorepSSHPublicKey, error) {
	resp, e := store.Get(context.TODO(), georepSSHKeysPrefix+volname, clientv3.WithPrefix())
	if e != nil {
		log.WithError(e).WithField("volname", volname).Error("Couldn't retrive SSH Key from the node")
		return nil, e
	}

	if resp.Count < 1 {
		return nil, errors.New("SSH Public Keys not found")
	}

	sshkeys := make([]georepapi.GeorepSSHPublicKey, resp.Count)
	for idx, kv := range resp.Kvs {
		var sshkey georepapi.GeorepSSHPublicKey
		if e = json.Unmarshal(kv.Value, &sshkey); e != nil {
			log.WithError(e).Error("Failed to unmarshal the data into georepsshpubkey object")
			return nil, e
		}
		sshkeys[idx] = sshkey
	}
	return sshkeys, nil
}
