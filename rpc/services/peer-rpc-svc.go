package services

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
)

type PeerService int

var (
	opRet        int32
	opError      string
	etcdConfDir  = "/var/lib/glusterd/"
	etcdConfFile = etcdConfDir + "etcdenv.conf"
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

// storeETCDEnv will store etcd environment variable in etcdConfFile
// (/var/lib/glusterd/etcdenv.conf) it will be useful at the time of restarting
func storeETCDEnv(env *RPCEtcdEnvReq) error {
	utils.InitDir(etcdConfDir)
	fp, err := os.OpenFile(etcdConfFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdConfFile,
		}).Error("Failed to open etcdConfFile")
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString("ETCD_NAME=" + *env.Name + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdConfFile,
		}).Error("Failed to write Environment variable to etcdConfFile")
		return err
	}

	if _, err = fp.WriteString("ETCD_INITIAL_CLUSTER=" + *env.InitialCluster + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdConfFile,
		}).Error("Failed to write Environment variable to etcdConfFile")
		return err
	}

	if _, err = fp.WriteString("ETCD_INITIAL_CLUSTER_STATE=" + *env.ClusterState + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdConfFile,
		}).Error("Failed to write Environment variable to etcdConfFile")
		return err
	}

	return nil
}

// ExportAndStoreETCDEnv will export etcd environment variable
func (etcd *PeerService) ExportAndStoreETCDEnv(env *RPCEtcdEnvReq, reply *RPCEtcdEnvResp) error {
	opRet = 0
	opError = ""

	// Exporting etcd environment variable
	os.Setenv("ETCD_NAME", *env.Name)
	os.Setenv("ETCD_INITIAL_CLUSTER", *env.InitialCluster)
	os.Setenv("ETCD_INITIAL_CLUSTER_STATE", *env.ClusterState)

	// Storing etcd envioronment variable in
	// etcdConfFile (/var/lib/glusterd/etcdenv.conf) locally. So that upon
	// glusterd restart we can restore these environment variable again
	err := storeETCDEnv(env)
	if err != nil {
		opRet = -1
		opError = fmt.Sprintf("Could not able to write etcd environment variable. Aborting")
		log.WithField("error", err.Error()).Error("Could not able to write etcd environment variable. Aborting")
	}

	// Restarting etcd daemon
	etcdCtx, err := etcdmgmt.ReStartETCD()
	if err != nil {
		opRet = -1
		opError = fmt.Sprintf("Could not able to restart etcd at remote node")
		log.WithField("error", err.Error()).Error("Could not able to restart etcd")
	}
	context.EtcdProcessCtx = etcdCtx

	reply.OpRet = &opRet
	reply.OpError = &opError

	return nil
}
