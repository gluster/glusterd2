package transaction

import (
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/peer"

	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
	netctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RunStepOn(step string, node uuid.UUID, c TxnCtx) (TxnCtx, error) {
	// TODO: I'm creating connections on demand. This should be changed so that
	// we have long term connections.
	p, err := peer.GetPeerF(node.String())
	if err != nil {
		return nil, err
	}

	var conn *grpc.ClientConn
	port := config.GetString("rpcport")

	for _, addr := range p.Addresses {
		conn, err = grpc.Dial(addr + ":" + port)
		if err == nil && conn != nil {
			break
		}
	}
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	client := NewTxnSvcClient(conn)

	req := &TxnStepReq{
		StepFunc: step,
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	req.Context = data

	var rsp *TxnStepResp

	rsp, err = client.RunStep(netctx.TODO(), req)
	if err != nil {
		return nil, err
	}

	if rsp.Error != "" {
		return nil, errors.New(rsp.Error)
	}

	rspCtx := new(txnCtx)
	err = json.Unmarshal(rsp.Resp, rspCtx)

	return rspCtx, err
}
