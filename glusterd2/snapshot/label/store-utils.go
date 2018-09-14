package label

import (
	"context"
	"encoding/json"

	gdstore "github.com/gluster/glusterd2/glusterd2/store"
	gderror "github.com/gluster/glusterd2/pkg/errors"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

const (
	labelPrefix string = "labels/"
)

var (
	//ExistsFunc check whether a given label exist or not
	ExistsFunc = Exists
	// AddOrUpdateLabelFunc marshals to label object and passes to store to add/update
	AddOrUpdateLabelFunc = AddOrUpdateLabel
)

//Exists check whether a given label exist or not
func Exists(name string) bool {
	resp, e := gdstore.Get(context.TODO(), labelPrefix+name)
	if e != nil {
		return false
	}

	return resp.Count == 1
}

//GetLabels retrives the json objects from the store and converts them into
//respective Info objects
func GetLabels() ([]*Info, error) {
	resp, e := gdstore.Get(context.TODO(), labelPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	labels := make([]*Info, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		var label Info

		if err := json.Unmarshal(kv.Value, &label); err != nil {
			log.WithError(err).WithField("Label", string(kv.Key)).Error("Failed to unmarshal label")
			continue
		}

		labels[i] = &label
	}

	return labels, nil
}

// AddOrUpdateLabel marshals to Label object and passes to store to add/update
func AddOrUpdateLabel(labelInfo *Info) error {
	json, e := json.Marshal(labelInfo)
	if e != nil {
		log.WithError(e).Error("Failed to marshal the labelinfo object")
		return e
	}

	_, e = gdstore.Put(context.TODO(), GetStorePath(labelInfo), string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add label to store")
		return e
	}
	return nil
}

// GetLabel fetches the json object from the store and unmarshalls it into
// Info object
func GetLabel(name string) (*Info, error) {
	var labelinfo Info

	resp, e := gdstore.Get(context.TODO(), labelPrefix+name)
	if e != nil {
		log.WithError(e).Error("Couldn't retrive volume from store")
		return nil, e
	}

	if resp.Count != 1 {
		log.WithField("label", name).Error("label not found")
		return nil, gderror.ErrLabelNotFound
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &labelinfo); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into labelinfo object")
		return nil, e
	}
	return &labelinfo, nil
}

//DeleteLabel passes the label path to store to delete the label object
func DeleteLabel(labelInfo *Info) error {
	_, e := gdstore.Delete(context.TODO(), GetStorePath(labelInfo))
	if e != nil {
		return e
	}

	/*
		TODO
		Delete all object tagged to this label
	*/
	return e
}

//GetStorePath return label path for etcd store
func GetStorePath(labelInfo *Info) string {
	return labelPrefix + labelInfo.Name
}
