package traceutils

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/store"
	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
)

const (
	traceMgmtPrefix string = "tracemgmt/"
)

// GetTraceConfig fetches the json object from the store and unmarshalls it
// into the JaegerConfigInfo object
func GetTraceConfig() (*tracemgmtapi.JaegerConfigInfo, error) {
	var t tracemgmtapi.JaegerConfigInfo
	resp, e := store.Get(context.TODO(), traceMgmtPrefix)
	if e != nil {
		return nil, e
	}

	if resp.Count != 1 {
		return nil, errors.New("trace config not found")
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &t); e != nil {
		return nil, e
	}
	return &t, nil
}

// AddOrUpdateTraceConfig marshals the JaegerConfigInfo object and saves it
// to the store
func AddOrUpdateTraceConfig(t *tracemgmtapi.JaegerConfigInfo) error {
	json, err := json.Marshal(t)
	if err != nil {
		return err
	}

	_, err = store.Put(context.TODO(), traceMgmtPrefix, string(json))
	return err
}

// DeleteTraceConfig deletes the JagegerConfigInfo object from the store
func DeleteTraceConfig() error {
	_, err := store.Delete(context.TODO(), traceMgmtPrefix)
	return err
}
