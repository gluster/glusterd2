package transaction

import (
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/rpc/server"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type txnSvc int

func init() {
	server.Register(new(txnSvc))
}

// RunStep handles the incoming request. It executes the requested step and returns the results
func (p *txnSvc) RunStep(rpcCtx context.Context, req *TxnStepReq) (*TxnStepResp, error) {
	var ctx txnCtx

	err := json.Unmarshal(req.Context, &ctx)
	if err != nil {
		log.WithError(err).Error("failed to Unmarshal transaction context")
		return nil, err
	}

	ctx.Logger().WithField("stepfunc", req.StepFunc).Debug("RunStep request recieved")

	f, ok := GetStepFunc(req.StepFunc)
	if !ok {
		log.WithField("stepfunc", req.StepFunc).Error("step function not found in registry")
		return nil, errors.New("step function not found")
	}

	ctx.Logger().WithField("stepfunc", req.StepFunc).Debug("running step")

	resp := new(TxnStepResp)

	err = f(&ctx)
	if err != nil {
		resp.Error = err.Error()
	} else {
		b, err := json.Marshal(ctx)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Resp = b
		}
	}

	return resp, nil
}

// RegisterService registers txnSvc with the given grpc.Server
func (p *txnSvc) RegisterService(s *grpc.Server) {
	RegisterTxnSvcServer(s, p)
}
