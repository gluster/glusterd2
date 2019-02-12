package transaction

import (
	"context"

	"go.opencensus.io/trace"
)

func newtracingExecutor(next Executor) Executor {
	return &tracingExecutor{next}
}

// tracingExecutor is middleware to record trace info.
// It will record the trace info and execute the next executor
type tracingExecutor struct {
	next Executor
}

// Execute will record trace info for Execute operation.
func (t *tracingExecutor) Execute(ctx context.Context, txn *Txn) error {
	ctx, span := trace.StartSpan(ctx, "txnEng.executor.Execute/")
	defer span.End()
	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
	)
	return t.next.Execute(ctx, txn)
}

// Resume will record trace info for Resume operation.
func (t *tracingExecutor) Resume(ctx context.Context, txn *Txn) error {
	ctx, span := trace.StartSpan(ctx, "txnEng.executor.Resume/")
	defer span.End()
	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
	)
	return t.next.Resume(ctx, txn)
}
