package transaction

import (
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/servers/peerrpc"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type txnSvc int

func init() {
	peerrpc.Register(new(txnSvc))
}

// RunStep handles the incoming request. It executes the requested step and returns the results
func (p *txnSvc) RunStep(rpcCtx context.Context, req *TxnStepReq) (*TxnStepResp, error) {
	var ctx Tctx

	err := json.Unmarshal(req.Context, &ctx)
	if err != nil {
		log.WithError(err).Error("failed to Unmarshal transaction context")
		return nil, err
	}

	logger := ctx.Logger().WithField("stepfunc", req.StepFunc)
	logger.Debug("RunStep request received")

	f, ok := GetStepFunc(req.StepFunc)
	if !ok {
		logger.Error("step function not found in registry")
		return nil, errors.New("step function not found")
	}

	logger.Debug("running step")

	resp := new(TxnStepResp)

	// Execute the step function, build and return result
	err = f(&ctx)
	if err != nil {
		logger.WithError(err).Debug("step function failed")
		resp.Error = err.Error()
	} else {
		b, err := json.Marshal(ctx)
		if err != nil {
			logger.WithError(err).Debug("failed to JSON marshal transcation context")
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
