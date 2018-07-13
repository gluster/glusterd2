package volgen

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
	_, err := store.Put(context.TODO(), volfilePrefix+name, content)
	return err
}

// GetVolfiles returns list of all Volfiles
func GetVolfiles() ([]string, error) {
	resp, e := store.Get(context.TODO(), volfilePrefix, clientv3.WithPrefix())
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
	resp, e := store.Get(context.TODO(), volfile, clientv3.WithPrefix())
	if e != nil {
		return []byte{}, errors.ErrFetchingVolfileContent
	}
	if len(resp.Kvs) == 0 {
		return []byte{}, errors.ErrVolFileNotFound
	}
	return resp.Kvs[0].Value, nil
}

// DeleteVolfiles deletes all the Volfiles with given prefix
func DeleteVolfiles(prefix string) error {
	_, err := store.Delete(context.TODO(), volfilePrefix+prefix, clientv3.WithPrefix())
	return err
}
