package transaction

import (
	"context"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/transaction"

	"go.opencensus.io/trace"
)

func newTracingManager(next StepManager) StepManager {
	return &tracingManager{next}
}

type tracingManager struct {
	next StepManager
}

// RunStep is a middleware which creates tracing span for step.DoFunc
func (t *tracingManager) RunStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) (err error) {
	spanName := fmt.Sprintf("RunStep/%s", step.DoFunc)
	ctx, span := trace.StartSpan(ctx, spanName)
	defer span.End()

	defer func() {
		attrs := []trace.Attribute{
			trace.StringAttribute("reqID", txnCtx.GetTxnReqID()),
		}
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
		}
		span.AddAttributes(attrs...)
	}()

	return t.next.RunStep(ctx, step, txnCtx)
}

// RollBackStep is a middleware which creates tracing span for step.UndoFunc
func (t *tracingManager) RollBackStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) (err error) {
	spanName := fmt.Sprintf("RollBackStep/%s", step.UndoFunc)
	ctx, span := trace.StartSpan(ctx, spanName)
	defer span.End()

	defer func() {
		attrs := []trace.Attribute{
			trace.StringAttribute("reqID", txnCtx.GetTxnReqID()),
		}
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
		}
		span.AddAttributes(attrs...)
	}()

	return t.next.RollBackStep(ctx, step, txnCtx)
}

// SyncStep is a middleware which creates tracing span for Sync steps
func (t *tracingManager) SyncStep(ctx context.Context, stepIndex int, txn *Txn) (err error) {
	spanName := fmt.Sprintf("SyncStep/%s", txn.Steps[stepIndex].DoFunc)
	ctx, span := trace.StartSpan(ctx, spanName)
	defer span.End()

	defer func() {
		attrs := []trace.Attribute{
			trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		}
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
		}
		span.AddAttributes(attrs...)
	}()

	return t.next.SyncStep(ctx, stepIndex, txn)
}
