package etcdmgmt

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/clientv3"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	config "github.com/spf13/viper"
	etcdcontext "golang.org/x/net/context"
)

// initEtcdClient will initialize etcd client. This instance of the client
// is only used to maintain/modify etcd cluster membership. For storing
// key-values in etcd store, libkv is used instead.
func initEtcdClient() (*etcd.Client, error) {

	endpoint := "http://" + config.GetString("etcdclientaddress")

	cfg := etcd.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	}

	client, err := etcd.New(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// EtcdMemberList returns a list of members in etcd cluster.
func EtcdMemberList() ([]*pb.Member, error) {

	client, err := initEtcdClient()
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberList: Failed to create etcd client.")
		return nil, err
	}
	defer client.Close()

	resp, err := client.MemberList(etcdcontext.Background())
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberList: Failed to list etcd members.")
		return nil, err
	}

	return resp.Members, nil
}

// EtcdMemberAdd will add a new member to the etcd cluster.
func EtcdMemberAdd(peerURL string) (*pb.Member, error) {

	client, err := initEtcdClient()
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberAdd: Failed to create etcd client.")
		return nil, err
	}
	defer client.Close()

	resp, err := client.MemberAdd(etcdcontext.Background(), []string{peerURL})
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberAdd: Failed to add etcd member.")
		return nil, err
	}

	return resp.Member, nil
}

// EtcdMemberRemove will remove a member from the etcd cluster.
func EtcdMemberRemove(memberID uint64) error {

	client, err := initEtcdClient()
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberRemove: Failed to create etcd client.")
		return err
	}
	defer client.Close()

	_, err = client.MemberRemove(etcdcontext.Background(), memberID)
	if err != nil {
		log.WithField("error", err).Debug("EtcdMemberRemove: Failed to remove etcd member.")
		return err
	}

	return nil
}

// EtcdMemberStatus will return status of etcd instance running on the node
// specified by the memberID
func EtcdMemberStatus(memberID uint64) (*etcd.StatusResponse, error) {

	var endpoint string

	mlist, err := EtcdMemberList()
	if err != nil {
		log.WithField("error", err).Debug("Failed to list members in etcd cluster")
		return nil, err
	}

	for _, m := range mlist {
		if m.ID == memberID {
			endpoint = m.ClientURLs[0]
		}
	}

	if endpoint == "" {
		return nil, errors.New("MemberID not found")
	}

	// Use a new client instance with member endpoint as client address.
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer client.Close()

	mapi := etcd.NewMaintenance(client)
	resp, err := mapi.Status(etcdcontext.Background(), endpoint)
	if err != nil {
		log.WithField("err", err).Debug("Failed to get member status.")
		return nil, err
	}

	return resp, nil
}
