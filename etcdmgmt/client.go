package etcdmgmt

import (
	"time"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
)

var (
	etcdClient etcd.Client
)

// InitEtcdClient will initialize etcd client. This instance of the client
// should only be used to maintain/modify cluster membership. For storing
// key-values in etcd store, one should use libkv instead.
func InitEtcdClient(endpoints string) error {
	cfg := etcd.Config{
		Endpoints:               []string{endpoints},
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: 3 * time.Second,
	}
	c, err := etcd.New(cfg)
	if err != nil {
		log.WithField("error", err).Error("Failed to initialize etcd client(v2).")
		return err
	}
	etcdClient = c
	return nil
}

// GetEtcdMemberAPI returns the etcd MemberAPI
func GetEtcdMembersAPI() etcd.MembersAPI {
	return etcd.NewMembersAPI(etcdClient)
}
