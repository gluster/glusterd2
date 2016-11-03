package peercommands

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/server"
	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	netctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// PeerService will be handling client requests on the server side for peer ops
type PeerService int

func init() {
	server.Register(new(PeerService))
}

// RegisterService registers a service
func (p *PeerService) RegisterService(s *grpc.Server) {
	RegisterPeerServiceServer(s, p)
}

var (
	etcdConfDir  = "/var/lib/glusterd/"
	etcdConfFile = etcdConfDir + "etcdenv.conf"
)

// ValidateAdd validates AddPeer operation at server side
func (p *PeerService) ValidateAdd(nc netctx.Context, args *PeerAddReq) (*PeerAddResp, error) {
	var opRet int32
	var opError string
	uuid := gdctx.MyUUID.String()

	if gdctx.MaxOpVersion < 40000 {
		opRet = -1
		opError = fmt.Sprintf("GlusterD instance running on %s is not compatible", args.Name)
	}
	peers, _ := peer.GetPeersF()
	if len(peers) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s is already part of another cluster", args.Name)
	}
	volumes, _ := volume.GetVolumes()
	if len(volumes) != 0 {
		opRet = -1
		opError = fmt.Sprintf("Peer %s already has existing volumes", args.Name)
	}

	reply := &PeerAddResp{
		OpRet:   opRet,
		OpError: opError,
		UUID:    uuid,
	}
	return reply, nil
}

// ValidateDelete validates DeletePeer operation at server side
func (p *PeerService) ValidateDelete(nc netctx.Context, args *PeerDeleteReq) (*PeerGenericResp, error) {
	var opRet int32
	var opError string
	// TODO : Validate if this guy has any volume configured where the brick(s) is
	// hosted in some other node, in that case the validation should fail

	reply := &PeerGenericResp{
		OpRet:   opRet,
		OpError: opError,
	}
	return reply, nil
}

// storeETCDEnv will store etcd environment in etcdenv config file
func storeETCDEnv(env *EtcdConfigReq) error {
	utils.InitDir(etcdmgmt.ETCDConfDir)
	fp, err := os.OpenFile(etcdmgmt.ETCDEnvFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDEnvFile,
		}).Error("Failed to open etcd env file")
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString("ETCD_NAME=" + env.Name + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDEnvFile,
			"key":   "ETCD_NAME",
			"val":   env.Name,
		}).Error("Failed to write Environment variable in to etcd conf file")
		return err
	}

	if _, err = fp.WriteString("ETCD_INITIAL_CLUSTER=" + env.InitialCluster + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDEnvFile,
			"key":   "ETCD_INITIAL_CLUSTER",
			"val":   env.InitialCluster,
		}).Error("Failed to write Environment variable in to etcd conf file")
		return err
	}

	if _, err = fp.WriteString("ETCD_INITIAL_CLUSTER_STATE=" + env.ClusterState + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDEnvFile,
			"key":   "ETCD_INITIAL_CLUSTER_STATE",
			"val":   env.ClusterState,
		}).Error("Failed to write Environment variable in to etcd conf file")
		return err
	}
	return nil
}

// storeETCDProxyConf will store etcd configuration for proxy etcd
func storeETCDProxyConf(env *EtcdConfigReq) error {
	utils.InitDir(etcdmgmt.ETCDConfDir)
	fp, err := os.OpenFile(etcdmgmt.ETCDProxyFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDProxyFile,
		}).Error("Failed to open etcd proxy file")
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString("ETCD_INITIAL_CLUSTER=" + env.InitialCluster + "\n"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  etcdmgmt.ETCDProxyFile,
			"key":   "ETCD_INITIAL_CLUSTER",
			"val":   env.InitialCluster,
		}).Error("Failed to write configuration in to etcd proxy file")
		return err
	}
	return nil
}

// ExportAndStoreETCDConfig will store & export etcd environment variable along
// with storing etcd configuration
func (p *PeerService) ExportAndStoreETCDConfig(nc netctx.Context, c *EtcdConfigReq) (*PeerGenericResp, error) {
	var opRet int32
	var opError string

	if c.Client == false {
		// Exporting etcd environment variable
		os.Setenv("ETCD_NAME", c.Name)
		os.Setenv("ETCD_INITIAL_CLUSTER", c.InitialCluster)
		os.Setenv("ETCD_INITIAL_CLUSTER_STATE", c.ClusterState)

		// Storing etcd envioronment variable in
		// etcdEnvFile (/var/lib/glusterd/etcdenv.conf) locally. So that upon
		// glusterd restart we can restore these environment variable again
		err := storeETCDEnv(c)
		if err != nil {
			opRet = -1
			opError = fmt.Sprintf("Could not able to write etcd configuration")
			log.WithField("error", err.Error()).Error("Could not able to write etcd configuration")
			return nil, err
		}
	} else {
		err := storeETCDProxyConf(c)
		if err != nil {
			opRet = -1
			opError = fmt.Sprintf("Could not able to write etcd proxy configuration")
			log.WithField("error", err.Error()).Error("Could not able to write etcd proxy configuration")
			return nil, err
		}
	}
	// Restarting etcd daemon
	etcdCtx, err := etcdmgmt.ReStartETCD()
	if err != nil {
		opRet = -1
		opError = fmt.Sprintf("Could not able to restart etcd at remote node")
		log.WithField("error", err.Error()).Error("Could not able to restart etcd")
	}
	gdctx.EtcdProcessCtx = etcdCtx

	reply := &PeerGenericResp{
		OpRet:   opRet,
		OpError: opError,
	}

	return reply, nil
}
