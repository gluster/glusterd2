package services

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

type PeerService int
type ExportEtcd int

var (
	opRet        int32
	opError      string
	etcdConfDir  = "/var/lib/glusterd/"
	etcdConfFile = etcdConfDir + "etcdConf"
)

// Validate function checks all validation for AddPeer at server side
func (p *PeerService) ValidateAdd(args *RPCPeerAddReq, reply *RPCPeerAddResp) error {
	opRet = 0
	opError = ""
	if context.MaxOpVersion < 40000 {
		opRet = -1
		opError = fmt.Sprintf("GlusterD instance running on %s is not compatible", *args.Name)
	}
	peers, _ := peer.GetPeers()
	if len(peers) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s is already part of another cluster", *args.Name)
	}
	volumes, _ := volume.GetVolumes()
	if len(volumes) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s already has existing volumes", *args.Name)
	}

	reply.OpRet = &opRet
	reply.OpError = &opError

	return nil
}

func storeEtcdEnv(env *RPCEtcdEnvReq) error {
	utils.InitDir(etcdConfDir)
	if err := ioutil.WriteFile(etcdConfFile, []byte("ETCD_NAME="+*env.Name), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(etcdConfFile, []byte("ETCD_INITIAL_CLUSTER="+*env.InitialCluster), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(etcdConfFile, []byte("ETCD_INITIAL_CLUSTER_STATE"+*env.ClusterState), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (etcd *ExportEtcd) ExportAndStoreEtcdEnv(env *RPCEtcdEnvReq, reply *RPCEtcdEnvResp) error {
	opRet = 0
	opError = ""

	if context.MaxOpVersion < 40000 {
		opRet = -1
		opError = fmt.Sprintf("GlusterD instance running on %s is not compatible", *env.PeerName)
	}

	// Exporting etcd environment variable
	os.Setenv("ETCD_NAME", *env.Name)
	os.Setenv("ETCD_INITIAL_CLUSTER", *env.InitialCluster)
	os.Setenv("ETCD_INITIAL_CLUSTER_STATE", *env.ClusterState)

	// Storing there envioronment variable locally. So that upon glusterd
	// restart we can set these environment variable again
	err := storeEtcdEnv(env)
	if err != nil {
		opRet = -1
		opError = fmt.Sprintf("Could not able to write etcd environment variable. Aborting")
		log.WithField("error", err.Error()).Error("Could not able to write etcd environment variable. Aborting")
	}

	// Restarting etcd daemon
	etcdCmd, err := etcdmgmt.ReStartEtcd()
	if err != nil {
		opRet = -1
		opError = fmt.Sprintf("Could not able to restart etcd at remote node")
		log.WithField("error", err.Error()).Error("Could not able to restart etcd")
	}
	context.Init()
	context.EtcdProcessCtx = etcdCmd

	reply.OpRet = &opRet
	reply.OpError = &opError

	return nil
}
