package transaction

import (
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/context"

	log "github.com/Sirupsen/logrus"
)

type TxnSvc int

func (p *TxnSvc) RunStep(req *TxnStepReq, resp *TxnStepResp) error {
	var ctx context.Context

	err := json.Unmarshal(req.Context, &ctx)
	if err != nil {
		log.WithError(err).Error("failed to Unmarshal transaction context")
		return err
	}

	ctx.Log.WithField("stepfunc", *req.StepFunc).Debug("RunStep request recieved")

	f, ok := GetStepFunc(*req.StepFunc)
	if !ok {
		log.WithField("stepfunc", *req.StepFunc).Error("step function not found in registry")
		return errors.New("step function not found")
	}

	ctx.Log.WithField("stepfunc", *req.StepFunc).Debug("running step")

	err = f(&ctx)
	if err != nil {
		*resp.Error = err.Error()
	} else {
		b, err := json.Marshal(ctx)
		if err != nil {
			*resp.Error = err.Error()
		} else {
			resp.Resp = b
		}
	}

	return nil
}
