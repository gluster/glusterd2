package volgen2

import (
	"context"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
)

var (
	volfilePrefix = "volfiles/"
)

func save(name string, content string) error {
	if _, err := store.Store.Put(context.TODO(), volfilePrefix+name, content); err != nil {
		return err
	}
	return nil
}

// GetVolfiles returns list of all Volfiles
func GetVolfiles() ([]string, error) {
	resp, e := store.Store.Get(context.TODO(), volfilePrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	volfiles := make([]string, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		volfiles[i] = string(kv.Key)
	}

	return volfiles, nil
}
