package transaction

import (
	"encoding/json"
	"errors"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/peer"

	"github.com/kshlm/pbrpc/pbcodec"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

func RunStepOn(step string, node uuid.UUID, c TxnCtx) (TxnCtx, error) {
	// TODO: I'm creating connections on demand. This should be changed so that
	// we have long term connections.
	p, err := peer.GetPeerF(node.String())
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	port := config.GetString("rpcport")

	for _, addr := range p.Addresses {
		conn, err = net.Dial("tcp", addr+":"+port)
		if err == nil && conn != nil {
			break
		}
	}
	if conn == nil {
		return nil, err
	}

	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(conn))
	defer client.Close()

	req := &TxnStepReq{
		StepFunc: new(string),
	}
	*req.StepFunc = step
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	req.Context = data

	rsp := new(TxnStepResp)

	err = client.Call("TxnSvc.RunStep", req, rsp)
	if err != nil {
		return nil, err
	}

	if *rsp.Error != "" {
		return nil, errors.New(*rsp.Error)
	}

	rspCtx := new(txnCtx)
	err = json.Unmarshal(rsp.Resp, rspCtx)

	return rspCtx, err
}
