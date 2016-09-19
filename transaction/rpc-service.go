package transaction

import (
	"encoding/json"
	"errors"

	log "github.com/Sirupsen/logrus"
)

type TxnSvc int

func (p *TxnSvc) RunStep(req *TxnStepReq, resp *TxnStepResp) error {
	var ctx txnCtx

	err := json.Unmarshal(req.Context, &ctx)
	if err != nil {
		log.WithError(err).Error("failed to Unmarshal transaction context")
		return err
	}

	ctx.Logger().WithField("stepfunc", *req.StepFunc).Debug("RunStep request recieved")

	f, ok := GetStepFunc(*req.StepFunc)
	if !ok {
		log.WithField("stepfunc", *req.StepFunc).Error("step function not found in registry")
		return errors.New("step function not found")
	}

	ctx.Logger().WithField("stepfunc", *req.StepFunc).Debug("running step")

	resp.Error = new(string)
	resp.Resp = make([]byte, 0)

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
