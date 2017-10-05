package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
)

const (
	daemonsPrefix = store.GlusterPrefix + "daemons/"
)

// save saves the daemon information in the store
func saveDaemon(d Daemon) error {
	p := path.Join(daemonsPrefix, gdctx.MyUUID.String(), d.ID())

	sd := newStoredDaemon(d)
	data, err := json.Marshal(sd)
	if err != nil {
		return err
	}

	_, err = store.Store.Put(context.TODO(), p, string(data))
	return err
}

func delDaemon(d Daemon) error {
	p := path.Join(daemonsPrefix, gdctx.MyUUID.String(), d.ID())

	_, err := store.Store.Delete(context.TODO(), p)

	return err
}

func getDaemon(id string) (Daemon, error) {
	p := path.Join(daemonsPrefix, gdctx.MyUUID.String(), id)

	resp, err := store.Store.Get(context.TODO(), p)
	if err != nil {
		return nil, err
	}

	if resp.Count != 1 {
		return nil, errors.New("daemon not found")
	}

	return unmarshalStoredDaemon(resp.Kvs[0].Value)
}

func getDaemons() ([]Daemon, error) {
	p := path.Join(daemonsPrefix, gdctx.MyUUID.String())

	resp, err := store.Store.Get(context.TODO(), p, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var ds []Daemon

	for _, kv := range resp.Kvs {
		d, err := unmarshalStoredDaemon(kv.Value)
		if err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}

	return ds, nil
}

func unmarshalStoredDaemon(data []byte) (*storedDaemon, error) {
	var sd storedDaemon
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}
