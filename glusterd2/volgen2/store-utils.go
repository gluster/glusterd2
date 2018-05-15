package volgen2

import (
	"context"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/errors"

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
		volFile := strings.TrimPrefix(string(kv.Key), "volfiles/")
		volfiles[i] = volFile
	}

	return volfiles, nil
}

//GetVolfile return particular volfile info
func GetVolfile(volfileID string) ([]byte, error) {
	volfile := volfilePrefix + volfileID
	resp, e := store.Store.Get(context.TODO(), volfile, clientv3.WithPrefix())
	if e != nil {
		return []byte{}, e
	}
	if len(resp.Kvs) == 0 {
		return []byte{}, errors.ErrVolFileNotFound
	}
	return resp.Kvs[0].Value, nil
}
