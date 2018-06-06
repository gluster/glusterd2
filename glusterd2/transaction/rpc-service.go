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

	var (
		resp   TxnStepResp
		f      StepFunc
		err    error
		ok     bool
		logger log.FieldLogger
	)

	var ctx Tctx
	if err = json.Unmarshal(req.Context, &ctx); err != nil {
		log.WithError(err).Error("failed to Unmarshal transaction context")
		goto End
	}

	logger = ctx.Logger().WithField("stepfunc", req.StepFunc)
	logger.Debug("RunStep request received")

	f, ok = getStepFunc(req.StepFunc)
	if !ok {
		err = errors.New("step function not found in registry")
		goto End
	}

	logger.Debug("executing step function")
	if err = f(&ctx); err != nil {
		logger.WithError(err).Error("step function failed")
		goto End
	}

	if err = ctx.commit(); err != nil {
		logger.WithError(err).Error("failed to commit txn context to store")
	}

End:
	// Ensure RPC will always send a success reply. Error is stored in
	// body of response.
	if err != nil {
		resp.Error = err.Error()
	}

	return &resp, nil
}

// RegisterService registers txnSvc with the given grpc.Server
func (p *txnSvc) RegisterService(s *grpc.Server) {
	RegisterTxnSvcServer(s, p)
}
