package transaction

import (
	"context"
	"errors"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/pborman/uuid"
	"go.opencensus.io/trace"
)

// Executor contains method set to execute a txn on local node
type Executor interface {
	Execute(ctx context.Context, txn *Txn) error
	IsTxnPending(txnID uuid.UUID) bool
}

// NewExecutor returns an Executor instance
func NewExecutor() Executor {
	e := &executorImpl{
		txnManager:  NewTxnManager(store.Store.Watcher),
		stepManager: newStepManager(),
		selfNodeID:  gdctx.MyUUID,
	}

	return e
}

// executorImpl is a concrete implementation of Executor
type executorImpl struct {
	txnManager  TxnManager
	stepManager StepManager
	selfNodeID  uuid.UUID
}

// IsTxnPending returns true if the given txn is in pending state
func (e *executorImpl) IsTxnPending(txnID uuid.UUID) bool {
	status, err := e.txnManager.GetTxnStatus(txnID, e.selfNodeID)
	if err != nil || status.State != txnPending {
		return false
	}
	return true
}

// Execute will run all steps of a given txn on local Node. If a step is marked as synchornized,
// then It will wait for all previous steps to complete on all involved Nodes.
// If a node is an initiator node then It will acquire all cluster locks before running the txn steps.
func (e *executorImpl) Execute(ctx context.Context, txn *Txn) error {
	var (
		errChan          = make(chan error)
		done             = make(chan struct{})
		updateStatusOnce = &sync.Once{}
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if !e.shouldRunTxn(txn) {
		return nil
	}

	ctx, span := trace.StartSpan(ctx, "txnEng.executor.Execute/")
	defer span.End()
	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
	)

	txn.Ctx.Logger().Info("transaction started on node")

	failureChan := e.watchTxnForFailure(ctx, txn)

	for i := range txn.Steps {
		go e.runTxnStep(ctx, i, txn, errChan, done)

		select {
		case err := <-errChan:
			status := TxnStatus{State: txnFailed, Reason: err.Error(), TxnID: txn.ID}
			e.txnManager.UpDateTxnStatus(status, txn.ID, e.selfNodeID)
			return err

		case err := <-failureChan:
			txn.Ctx.Logger().WithError(err).Error("transaction got failed on some other node, cancelling ongoing txn on this node")
			return err

		case <-done:
			updateStatusOnce.Do(func() {
				e.txnManager.UpDateTxnStatus(TxnStatus{State: txnRunning, TxnID: txn.ID}, txn.ID, e.selfNodeID)
			})
		}
	}

	return e.txnManager.UpDateTxnStatus(TxnStatus{State: txnSucceeded, TxnID: txn.ID}, txn.ID, e.selfNodeID)
}

// runTxnStep will run a txn Step having given `stepIndex`. On successful completion of step it will notify on given `done` chan.
// In case of any error it will send notification of given `errChan`.
// If a step is marked as sync then it will wait for all previous steps to complete on all involved nodes.
func (e *executorImpl) runTxnStep(ctx context.Context, stepIndex int, txn *Txn, errChan chan<- error, done chan<- struct{}) {
	var (
		step   = txn.Steps[stepIndex]
		logger = txn.Ctx.Logger().WithField("stepname", step.DoFunc)
	)

	logger.Debug("running step func on node")

	// a synchronized step is executed only after all previous steps
	// have been completed successfully by all involved peers.
	if step.Sync {
		logger.Debug("synchronizing txn step")
		if err := e.stepManager.SyncStep(ctx, stepIndex, txn); err != nil {
			logger.WithError(err).Error("encounter an error in synchronizing txn step")
			errChan <- err
			return
		}
		logger.Debug("transaction got synchronized")
	}

	if err := e.stepManager.RunStep(ctx, step, txn.Ctx); err != nil {
		logger.WithError(err).Error("failed in executing txn step")
		e.stepManager.RollBackStep(ctx, step, txn.Ctx)
		errChan <- err
		return
	}

	logger.Debug("step func executed successfully on node")

	if err := e.txnManager.UpdateLastExecutedStep(stepIndex, txn.ID, e.selfNodeID); err != nil {
		logger.WithError(err).Error("failed in updating last executed step to store")
		e.stepManager.RollBackStep(ctx, step, txn.Ctx)
		errChan <- err
		return
	}

	done <- struct{}{}
}

// watchTxnForFailure will watch a given txn for failure. It returns an error chan.
// If the txn got marked as failure by some other node, then it will notify on returned chan.
func (e *executorImpl) watchTxnForFailure(ctx context.Context, txn *Txn) <-chan error {
	var (
		failureChan   = make(chan error)
		txnStatusChan = e.txnManager.WatchTxnStatus(ctx.Done(), txn.ID, e.selfNodeID)
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case status := <-txnStatusChan:
				if status.State == txnFailed {
					failureChan <- errors.New(status.Reason)
					return
				}
			}
		}
	}()

	return failureChan
}

// shouldRunTxn returns true if peer is involved in the txn
func (e *executorImpl) shouldRunTxn(txn *Txn) bool {
	for _, nodeID := range txn.Nodes {
		if uuid.Equal(nodeID, e.selfNodeID) {
			return true
		}
	}
	txn.Ctx.Logger().Debug("skipping txn on this node")
	return false
}
