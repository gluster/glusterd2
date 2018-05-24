package transaction

import (
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	netctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// runStepOn will run the step on the specified node
func runStepOn(step string, node uuid.UUID, c TxnCtx) error {
	// TODO: I'm creating connections on demand. This should be changed so that
	// we have long term connections.
	p, err := peer.GetPeerF(node.String())
	if err != nil {
		c.Logger().WithFields(log.Fields{
			"peerid": node.String(),
			"error":  err,
		}).Error("peer not found")
		return err
	}

	logger := c.Logger().WithField("remotepeer", p.ID.String()+"("+p.Name+")")

	var conn *grpc.ClientConn

	remote, err := utils.FormRemotePeerAddress(p.PeerAddresses[0])
	if err != nil {
		return err
	}

	conn, err = grpc.Dial(remote, grpc.WithInsecure())
	if err == nil && conn != nil {
		logger.WithFields(log.Fields{
			"remote": remote,
		}).Debug("connected to remote")
	}

	if conn == nil {
		logger.WithFields(log.Fields{
			"error":  err,
			"remote": p.PeerAddresses[0],
		}).Error("failed to grpc.Dial remote")
		return err
	}
	defer conn.Close()

	client := NewTxnSvcClient(conn)

	req := &TxnStepReq{
		StepFunc: step,
	}
	data, err := json.Marshal(c)
	if err != nil {
		logger.WithError(err).Error("failed to JSON marshal transaction context")
		return err
	}
	req.Context = data

	var rsp *TxnStepResp

	rsp, err = client.RunStep(netctx.TODO(), req)
	if err != nil {
		logger.WithFields(log.Fields{
			"error": err,
			"rpc":   "TxnSvc.RunStep",
		}).Error("failed RPC call")
		return err
	}

	if rsp.Error != "" {
		logger.WithError(errors.New(rsp.Error)).Error("TxnSvc.Runstep failed on peer")
		return errors.New(rsp.Error)
	}

	return nil
}
