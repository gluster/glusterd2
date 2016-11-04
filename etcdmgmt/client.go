package etcdmgmt

import (
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/clientv3"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	etcdcontext "golang.org/x/net/context"
)

var etcdClient struct {
	client *etcd.Client
	sync.Mutex
}

// InitEtcdClient will initialize etcd client. This instance of the client
// should only be used to maintain/modify cluster membership. For storing
// key-values in etcd store, one should use libkv instead.
func InitEtcdClient(endpoint string) error {
	cfg := etcd.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	}

	etcdClient.Lock()
	defer etcdClient.Unlock()

	if etcdClient.client != nil {
		return errors.New("An instance of etcd client is already active.")
	}

	c, err := etcd.New(cfg)
	if err != nil {
		return err
	}

	etcdClient.client = c
	return nil
}

// CloseEtcdClient shuts down the client's etcd connections. If the client is
// not closed, the connection will have leaky goroutines.
func CloseEtcdClient() error {
	etcdClient.Lock()
	defer etcdClient.Unlock()

	if etcdClient.client == nil {
		return errors.New("Etcd client is not initialized.")
	}

	err := etcdClient.client.Close()
	if err != nil {
		return err
	}
	etcdClient.client = nil

	return nil
}

// EtcdMemberList returns a list of members in etcd cluster.
func EtcdMemberList() ([]*pb.Member, error) {

	resp, err := etcdClient.client.MemberList(etcdcontext.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error":               err,
			"ClusterId(Response)": resp.Header.ClusterId,
			"MemberId(Response)":  resp.Header.MemberId,
			"Revision":            resp.Header.Revision,
			"RaftTerm":            resp.Header.RaftTerm,
		}).Debug("EtcdMemberList: Failed to list etcd members.")
		return nil, err
	}

	return resp.Members, nil
}

// EtcdMemberAdd will add a new member to the etcd cluster.
func EtcdMemberAdd(peerURL string) (*pb.Member, error) {

	resp, err := etcdClient.client.MemberAdd(etcdcontext.Background(), []string{peerURL})
	if err != nil {
		log.WithFields(log.Fields{
			"error":               err,
			"ClusterId(Response)": resp.Header.ClusterId,
			"MemberId(Response)":  resp.Header.MemberId,
			"Revision":            resp.Header.Revision,
			"RaftTerm":            resp.Header.RaftTerm,
		}).Debug("EtcdMemberAdd: Failed to add etcd member.")
		return nil, err
	}

	return resp.Member, nil
}

// EtcdMemberRemove will remove a member from the etcd cluster.
func EtcdMemberRemove(memberID uint64) error {

	resp, err := etcdClient.client.MemberRemove(etcdcontext.Background(), memberID)
	if err != nil {
		log.WithFields(log.Fields{
			"error":               err,
			"ClusterId(Response)": resp.Header.ClusterId,
			"MemberId(Response)":  resp.Header.MemberId,
			"Revision":            resp.Header.Revision,
			"RaftTerm":            resp.Header.RaftTerm,
		}).Debug("EtcdMemberRemove: Failed to remove etcd member.")
		return err
	}

	return nil
}
