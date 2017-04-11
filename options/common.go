package options

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"
)

const (
	clusterOptsPrefix string = store.GlusterPrefix + "clusteropts/"
	volumeOptsPrefix  string = store.GlusterPrefix + "volumeopts/"
)

// GetClusterOption gets cluster configuration from etcd store,
// returns default value if not available in store. Returns error if
// not available in defaults list.
func GetClusterOption(key string) (interface{}, error) {
	var v interface{}

	// Get Value from Store
	clusterKey := clusterOptsPrefix + key
	resp, e := gdctx.Store.Get(context.TODO(), clusterKey)
	if e == nil {
		if resp.Count == 1 {
			if e = json.Unmarshal(resp.Kvs[0].Value, &v); e != nil {
				return nil, e
			}
			return v, nil
		}
	}

	// If not available get value from defaults
	dv, dvOk := clusterDefaultOptions[key]
	if dvOk {
		return dv.defaultValue, nil
	}

	// Else return error
	return "", errors.New("Invalid Option")
}

// SetClusterOption validates and saves the cluster configuration in
// etcd store.
func SetClusterOption(key string, value interface{}) error {
	clusterKey := clusterOptsPrefix + key
	if clusterOptValidate(key, value) {
		d, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, e := gdctx.Store.Put(context.TODO(), clusterKey, string(d))
		return e
	}
	return errors.New("Invalid Value")
}

// ResetClusterOption deletes cluster configurations stored in etcd store.
func ResetClusterOption(key string) error {
	clusterKey := clusterOptsPrefix + key
	_, e := gdctx.Store.Delete(context.TODO(), clusterKey)
	return e
}

// GetVolumeOption gets volume configuration from etcd store,
// returns default value if not available in store. Returns error if
// not available in defaults list.
func GetVolumeOption(volname string, key string) (interface{}, error) {
	// Get Value from Store
	var v interface{}
	volumeKey := volumeOptsPrefix + volname + "/" + key
	resp, e := gdctx.Store.Get(context.TODO(), volumeKey)
	if e == nil {
		if resp.Count == 1 {
			if e = json.Unmarshal(resp.Kvs[0].Value, &v); e != nil {
				return nil, e
			}
			return v, nil
		}
	}

	// If not available get value from defaults
	dv, dvOk := volumeDefaultOptions[key]
	if dvOk {
		return dv.defaultValue, nil
	}

	// Else return error
	return "", errors.New("Invalid Option")
}

// SetVolumeOption validates and saves the volume configuration in
// etcd store.
func SetVolumeOption(volname string, key string, value interface{}) error {
	volumeKey := volumeOptsPrefix + volname + "/" + key
	if volumeOptValidate(key, value) {
		d, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, e := gdctx.Store.Put(context.TODO(), volumeKey, string(d))
		return e
	}
	return errors.New("Invalid Value")
}

// ResetVolumeOption deletes volume configurations stored in etcd store.
func ResetVolumeOption(volname string, key string) error {
	volumeKey := volumeOptsPrefix + volname + "/" + key
	_, e := gdctx.Store.Delete(context.TODO(), volumeKey)
	return e
}
